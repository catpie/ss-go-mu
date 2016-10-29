package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	. "github.com/catpie/ss-go-mu/log"
	"github.com/cyfdecyf/leakybuf"
	"github.com/orvice/shadowsocks-go/mu/user"
	ss "github.com/orvice/shadowsocks-go/shadowsocks"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	idType  = 0 // address type index
	idIP0   = 1 // ip addres start index
	idDmLen = 1 // domain address length index
	idDm0   = 2 // domain address start index

	typeIPv4 = 1 // type is ipv4 address
	typeDm   = 3 // type is domain address
	typeIPv6 = 4 // type is ipv6 address

	lenIPv4     = net.IPv4len + 2 // ipv4 + 2port
	lenIPv6     = net.IPv6len + 2 // ipv6 + 2port
	lenDmBase   = 2               // 1addrLen + 2port, plus addrLen
	lenHmacSha1 = 10
)

var ssdebug ss.DebugLog

func getRequest(conn *ss.Conn, auth bool) (host string, res_size int, ota bool, err error) {
	var n int
	ss.SetReadTimeout(conn)

	// buf size should at least have the same size with the largest possible
	// request size (when addrType is 3, domain name has at most 256 bytes)
	// 1(addrType) + 1(lenByte) + 256(max length address) + 2(port) + 10(hmac-sha1)
	buf := make([]byte, 270)
	// read till we get possible domain length field
	if n, err = io.ReadFull(conn, buf[:idType+1]); err != nil {
		return
	}
	res_size += n

	var reqStart, reqEnd int
	addrType := buf[idType]
	switch addrType & ss.AddrMask {
	case typeIPv4:
		reqStart, reqEnd = idIP0, idIP0+lenIPv4
	case typeIPv6:
		reqStart, reqEnd = idIP0, idIP0+lenIPv6
	case typeDm:
		if n, err = io.ReadFull(conn, buf[idType+1:idDmLen+1]); err != nil {
			return
		}
		reqStart, reqEnd = idDm0, int(idDm0+buf[idDmLen]+lenDmBase)
	default:
		err = fmt.Errorf("addr type %d not supported", addrType&ss.AddrMask)
		return
	}
	res_size += n

	if n, err = io.ReadFull(conn, buf[reqStart:reqEnd]); err != nil {
		return
	}
	res_size += n

	// Return string for typeIP is not most efficient, but browsers (Chrome,
	// Safari, Firefox) all seems using typeDm exclusively. So this is not a
	// big problem.
	switch addrType & ss.AddrMask {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}
	// parse port
	port := binary.BigEndian.Uint16(buf[reqEnd-2 : reqEnd])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	// if specified one time auth enabled, we should verify this
	if auth || addrType&ss.OneTimeAuthMask > 0 {
		ota = true
		if n, err = io.ReadFull(conn, buf[reqEnd:reqEnd+lenHmacSha1]); err != nil {
			return
		}
		iv := conn.GetIv()
		key := conn.GetKey()
		actualHmacSha1Buf := ss.HmacSha1(append(iv, key...), buf[:reqEnd])
		if !bytes.Equal(buf[reqEnd:reqEnd+lenHmacSha1], actualHmacSha1Buf) {
			err = fmt.Errorf("verify one time auth failed, iv=%v key=%v data=%v", iv, key, buf[:reqEnd])
			return
		}
		res_size += n
	}
	return
}

const logCntDelta = 100

var connCnt int
var nextLogConnCnt int = logCntDelta

func handleConnection(user UserInterface, conn *ss.Conn, auth bool) {
	var host string

	connCnt++ // this maybe not accurate, but should be enough
	if connCnt-nextLogConnCnt >= 0 {
		// XXX There's no xadd in the atomic package, so it's difficult to log
		// the message only once with low cost. Also note nextLogConnCnt maybe
		// added twice for current peak connection number level.
		Log.Debug("Number of client connections reaches %d\n", nextLogConnCnt)
		nextLogConnCnt += logCntDelta
	}

	// function arguments are always evaluated, so surround debug statement
	// with if statement
	Log.Debug(fmt.Sprintf("new client %s->%s\n", conn.RemoteAddr().String(), conn.LocalAddr()))
	closed := false
	defer func() {
		if ssdebug {
			Log.Debug(fmt.Sprintf("closed pipe %s<->%s\n", conn.RemoteAddr(), host))
		}
		connCnt--
		if !closed {
			conn.Close()
		}
	}()

	host, res_size, ota, err := getRequest(conn, auth)
	if err != nil {
		Log.Error("error getting request", conn.RemoteAddr(), conn.LocalAddr(), err)
		return
	}
	Log.Info(fmt.Sprintf("[port-%d]connecting %s ", user.GetPort(), host))
	remote, err := net.Dial("tcp", host)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			Log.Error("dial error:", err)
		} else {
			Log.Error("error connecting to:", host, err)
		}
		return
	}
	defer func() {
		if !closed {
			remote.Close()
		}
	}()

	// debug conn info
	Log.Debug(fmt.Sprintf("%d conn debug:  local addr: %s | remote addr: %s network: %s ", user.GetPort(),
		conn.LocalAddr().String(), conn.RemoteAddr().String(), conn.RemoteAddr().Network()))
	go func() {
		err = storage.IncrSize(user, res_size)
		if err != nil {
			Log.Error(err)
			return
		}
		err = storage.MarkUserOnline(user)
		if err != nil {
			Log.Error(err)
			return
		}
		Log.Debug(fmt.Sprintf("[port-%d] store size: %d", user.GetPort(), res_size))

		Log.Info(fmt.Sprintf("piping %s<->%s ota=%v connOta=%v", conn.RemoteAddr(), host, ota, conn.IsOta()))

	}()
	if ota {
		go PipeThenCloseOta(conn, remote, false, host, user)
	} else {
		go PipeThenClose(conn, remote, false, host, user)
	}

	PipeThenClose(remote, conn, true, host, user)
	closed = true
	return
}

type PortListener struct {
	password string
	listener net.Listener
}

type PasswdManager struct {
	sync.Mutex
	portListener map[string]*PortListener
}

func (pm *PasswdManager) add(port, password string, listener net.Listener) {
	pm.Lock()
	pm.portListener[port] = &PortListener{password, listener}
	pm.Unlock()
}

func (pm *PasswdManager) get(port string) (pl *PortListener, ok bool) {
	pm.Lock()
	pl, ok = pm.portListener[port]
	pm.Unlock()
	return
}

func (pm *PasswdManager) del(port string) {
	pl, ok := pm.get(port)
	if !ok {
		return
	}
	pl.listener.Close()
	pm.Lock()
	delete(pm.portListener, port)
	pm.Unlock()
}

var passwdManager = PasswdManager{portListener: map[string]*PortListener{}}

func waitSignal() {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	for sig := range sigChan {
		if sig == syscall.SIGHUP {
		} else {
			// is this going to happen?
			Log.Printf("caught signal %v, exit", sig)
			os.Exit(0)
		}
	}
}

func runWithCustomMethod(user UserInterface) {
	// port, password string, Cipher *ss.Cipher
	port := strconv.Itoa(user.GetPort())
	password := user.GetPasswd()
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		Log.Error(fmt.Sprintf("error listening port %v: %v\n", port, err))
		// os.Exit(1)
		return
	}
	passwdManager.add(port, password, ln)
	cipher, err, auth := user.GetCipher()
	if err != nil {
		return
	}
	Log.Info(fmt.Sprintf("server listening port %v ...\n", port))
	for {
		conn, err := ln.Accept()
		if err != nil {
			// listener maybe closed to update password
			Log.Debug(fmt.Sprintf("accept error: %v\n", err))
			return
		}
		// Creating cipher upon first connection.
		if cipher == nil {
			Log.Debug("creating cipher for port:", port)
			method := user.GetMethod()

			if strings.HasSuffix(method, "-auth") {
				method = method[:len(method)-5]
				auth = true
			} else {
				auth = false
			}

			cipher, err = ss.NewCipher(method, password)
			if err != nil {
				Log.Error(fmt.Sprintf("Error generating cipher for port: %s %v\n", port, err))
				conn.Close()
				continue
			}
		}
		go handleConnection(user, ss.NewConn(conn, cipher.Copy()), auth)
	}
}

const bufSize = 4096
const nBuf = 2048

func PipeThenClose(src, dst net.Conn, is_res bool, host string, user UserInterface) {
	var pipeBuf = leakybuf.NewLeakyBuf(nBuf, bufSize)
	defer dst.Close()
	buf := pipeBuf.Get()
	// defer pipeBuf.Put(buf)
	var size int

	for {
		SetReadTimeout(src)
		n, err := src.Read(buf)
		// read may return EOF with n > 0
		// should always process n > 0 bytes before handling error
		if n > 0 {
			size, err = dst.Write(buf[0:n])
			if is_res {
				go func() {
					err = storage.IncrSize(user, size)
					if err != nil {
						Log.Error(err)
					}
				}()
				Log.Debug(fmt.Sprintf("[port-%d] store size: %d", user.GetPort(), size))
			}
			if err != nil {
				Log.Debug("write:", err)
				break
			}
		}
		if err != nil || n == 0 {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			break
		}
	}
	return
}

func PipeThenCloseOta(src *ss.Conn, dst net.Conn, is_res bool, host string, user UserInterface) {
	const (
		dataLenLen  = 2
		hmacSha1Len = 10
		idxData0    = dataLenLen + hmacSha1Len
	)

	defer func() {
		dst.Close()
	}()
	var pipeBuf = leakybuf.NewLeakyBuf(nBuf, bufSize)
	buf := pipeBuf.Get()
	// sometimes it have to fill large block
	for i := 1; ; i += 1 {
		SetReadTimeout(src)
		n, err := io.ReadFull(src, buf[:dataLenLen+hmacSha1Len])
		if err != nil {
			if err == io.EOF {
				break
			}
			Log.Debug(fmt.Sprintf("conn=%p #%v read header error n=%v: %v", src, i, n, err))
			break
		}
		dataLen := binary.BigEndian.Uint16(buf[:dataLenLen])
		expectedHmacSha1 := buf[dataLenLen:idxData0]

		var dataBuf []byte
		if len(buf) < int(idxData0+dataLen) {
			dataBuf = make([]byte, dataLen)
		} else {
			dataBuf = buf[idxData0 : idxData0+dataLen]
		}
		if n, err := io.ReadFull(src, dataBuf); err != nil {
			if err == io.EOF {
				break
			}
			Log.Debug(fmt.Sprintf("conn=%p #%v read data error n=%v: %v", src, i, n, err))
			break
		}
		chunkIdBytes := make([]byte, 4)
		chunkId := src.GetAndIncrChunkId()
		binary.BigEndian.PutUint32(chunkIdBytes, chunkId)
		actualHmacSha1 := ss.HmacSha1(append(src.GetIv(), chunkIdBytes...), dataBuf)
		if !bytes.Equal(expectedHmacSha1, actualHmacSha1) {
			Log.Debug(fmt.Sprintf("conn=%p #%v read data hmac-sha1 mismatch, iv=%v chunkId=%v src=%v dst=%v len=%v expeced=%v actual=%v", src, i, src.GetIv(), chunkId, src.RemoteAddr(), dst.RemoteAddr(), dataLen, expectedHmacSha1, actualHmacSha1))
			break
		}

		if n, err := dst.Write(dataBuf); err != nil {
			Log.Debug(fmt.Sprintf("conn=%p #%v write data error n=%v: %v", dst, i, n, err))
			break
		}
		if is_res {
			go func() {
				err := storage.IncrSize(user, n)
				if err != nil {
					Log.Error(err)
				}
				Log.Debug(fmt.Sprintf("[port-%d] store size: %d", user.GetPort(), n))
			}()
		}
	}
	return
}

var readTimeout time.Duration

func SetReadTimeout(c net.Conn) {
	if readTimeout != 0 {
		c.SetReadDeadline(time.Now().Add(readTimeout))
	}
}

func showConn(raw_req_header, raw_res_header []byte, host string, user user.User, size int, is_http bool) {
	if size == 0 {
		Log.Error(fmt.Sprintf("[port-%d]  Error: request %s cancel", user.GetPort(), host))
		return
	}
	if is_http {
		req, _ := http.ReadRequest(bufio.NewReader(bytes.NewReader(raw_req_header)))
		if req == nil {
			lines := bytes.SplitN(raw_req_header, []byte(" "), 2)
			Log.Debug(fmt.Sprintf("%s http://%s/ \"Unknow\" HTTP/1.1 unknow user-port: %d size: %d\n", lines[0], host, user.GetPort(), size))
			return
		}
		res, _ := http.ReadResponse(bufio.NewReader(bytes.NewReader(raw_res_header)), req)
		statusCode := 200
		if res != nil {
			statusCode = res.StatusCode
		}
		Log.Debug(fmt.Sprintf("%s http://%s%s \"%s\" %s %d  user-port: %d  size: %d\n", req.Method, req.Host, req.URL.String(), req.Header.Get("user-agent"), req.Proto, statusCode, user.GetPort(), size))
	} else {
		Log.Debug(fmt.Sprintf("CONNECT %s \"NONE\" NONE NONE user-port: %d  size: %d\n", host, user.GetPort(), size))
	}

}

func checkHttp(extra []byte, conn *ss.Conn) (is_http bool, data []byte, err error) {
	var buf []byte
	var methods = []string{"GET", "HEAD", "POST", "PUT", "TRACE", "OPTIONS", "DELETE"}
	is_http = false
	if extra == nil || len(extra) < 10 {
		buf = make([]byte, 10)
		if _, err = io.ReadFull(conn, buf); err != nil {
			return
		}
	}

	if buf == nil {
		data = extra
	} else if extra == nil {
		data = buf
	} else {
		buffer := bytes.NewBuffer(extra)
		buffer.Write(buf)
		data = buffer.Bytes()
	}

	for _, method := range methods {
		if bytes.HasPrefix(data, []byte(method)) {
			is_http = true
			break
		}
	}
	return
}
