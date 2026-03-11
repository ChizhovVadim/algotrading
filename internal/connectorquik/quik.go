package connectorquik

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type QuikConnector struct {
	logger       *log.Logger
	port         int
	nextId       int64
	mu           sync.Mutex
	mainConn     net.Conn
	reader       *bufio.Reader
	writer       *transform.Writer
	callbackConn net.Conn
}

func New(
	logger *log.Logger,
	port int,
	startId int64,
) *QuikConnector {
	return &QuikConnector{
		logger: logger,
		port:   port,
		nextId: startId,
	}
}

func (q *QuikConnector) Init(
	ctx context.Context,
	callbackHandler func(context.Context, CallbackJson),
) error {
	mainConn, err := dial(q.port)
	if err != nil {
		return err
	}
	q.mainConn = mainConn

	callbackConn, err := dial(q.port + 1)
	if err != nil {
		return err
	}
	q.callbackConn = callbackConn

	var quikCharmap = charmap.Windows1251
	q.reader = bufio.NewReader(transform.NewReader(q.mainConn, quikCharmap.NewDecoder()))
	q.writer = transform.NewWriter(q.mainConn, quikCharmap.NewEncoder())

	// эта горутина завершатся, тк defer quik.Close() закроет callback connection.
	// даже если не хотим обрабатывать callbacks, то все равно нужно читать сообщения.
	go func() {
		var err = q.handleCallbacks(ctx, callbackHandler)
		if err != nil {
			if q.logger != nil {
				q.logger.Println("quik.handleCallbacks", "error", err)
			}
			return
		}
	}()
	return nil
}

func (q *QuikConnector) Close() error {
	var mainConnErr, callbackConnErr error
	if q.mainConn != nil {
		mainConnErr = q.mainConn.Close()
	}
	if q.callbackConn != nil {
		callbackConnErr = q.callbackConn.Close()
	}
	return errors.Join(mainConnErr, callbackConnErr)
}

func (q *QuikConnector) Execute(req RequestJson, resp *ResponseJson) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	req.Id = q.nextId
	q.nextId += 1
	req.CreatedTime = timeToQuikTime(time.Now())
	//TODO log clientName in request/response.
	if err := q.writeRequest(req); err != nil {
		return err
	}
	if err := q.readResponse(resp); err != nil {
		return err
	}
	if !(req.Id == resp.Id) {
		return fmt.Errorf("assert req.Id == resp.Id")
	}
	if resp.LuaError != "" {
		return fmt.Errorf("lua error: %v", resp.LuaError)
	}
	return nil
}

func (q *QuikConnector) writeRequest(request RequestJson) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}
	_, err = q.writer.Write(b)
	if err != nil {
		return err
	}
	_, err = q.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	if q.logger != nil {
		q.logger.Println(string(b))
	}
	return nil
}

func (q *QuikConnector) readResponse(resp *ResponseJson) error {
	incoming, err := q.reader.ReadString('\n')
	if err != nil {
		return err
	}
	if q.logger != nil {
		q.logger.Println(compactString(incoming, 2_048))
	}
	err = json.Unmarshal([]byte(incoming), resp)
	if err != nil {
		return err
	}
	return nil
}

func dial(port int) (net.Conn, error) {
	return net.Dial("tcp", "localhost:"+strconv.Itoa(port))
}

func timeToQuikTime(time time.Time) int64 {
	return time.UnixNano() / 1000
}

func (q *QuikConnector) MakeQuery(cmd string, data any) (ResponseJson, error) {
	var resp ResponseJson
	var err = q.Execute(RequestJson{
		Command: cmd,
		Data:    data,
	}, &resp)
	return resp, err
}

// TODO Можно логировать callbacks но не все!
func (q *QuikConnector) handleCallbacks(
	ctx context.Context,
	callbackHandler func(context.Context, CallbackJson),
) error {
	reader := bufio.NewReader(transform.NewReader(q.callbackConn, charmap.Windows1251.NewDecoder()))
	for {
		incoming, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		var callbackJson CallbackJson
		err = json.Unmarshal([]byte(incoming), &callbackJson)
		if err != nil {
			return err
		}
		if callbackHandler != nil {
			callbackHandler(ctx, callbackJson)
		}
	}
}

func compactString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	var sb = &strings.Builder{}
	for _, rune := range s {
		sb.WriteRune(rune)
		if sb.Len() >= maxLen {
			sb.WriteString("...")
			break
		}
	}
	return sb.String()
}
