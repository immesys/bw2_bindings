package bw2bind

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/cihub/seelog"
	"github.com/immesys/bw2/objects"
)

type BW2Client struct {
	c            net.Conn
	out          *bufio.Writer
	in           *bufio.Reader
	remotever    string
	seqnos       map[int]chan *Frame
	olock        sync.Mutex
	curseqno     uint32
	defAutoChain *bool
}

func (cl *BW2Client) GetSeqNo() int {
	newseqno := atomic.AddUint32(&cl.curseqno, 1)
	return int(newseqno)
}
func (cl *BW2Client) OverrideAutoChainTo(v bool) {
	cl.defAutoChain = &v
}
func (cl *BW2Client) ClearAutoChainOverride() {
	cl.defAutoChain = nil
}
func Connect(to string) (*BW2Client, error) {
	if to == "" {
		to = "localhost:28589"
	}
	_, _, err := net.SplitHostPort(to)
	if err != nil && err.Error() == "missing port in address" {
		to = to + ":28589"
		_, _, err = net.SplitHostPort(to)
	}
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", to)
	if err != nil {
		return nil, err
	}
	rv := &BW2Client{c: conn,
		out:    bufio.NewWriter(conn),
		in:     bufio.NewReader(conn),
		seqnos: make(map[int]chan *Frame)}

	//As a bit of a sanity check, we read the first frame, which is the
	//server HELO message
	ok := make(chan bool, 1)
	go func() {
		helo, err := LoadFrameFromStream(rv.in)
		if err != nil {
			log.Error("Malformed HELO frame: ", err)
			ok <- false
			return
		}
		if helo.Cmd != CmdHello {
			log.Error("Frame not HELO")
			ok <- false
			return
		}
		rver, hok := helo.GetFirstHeader("version")
		if !hok {
			log.Error("Frame has no version")
			ok <- false
			return
		}
		rv.remotever = rver
		log.Info("Connected to BOSSWAVE router version ", rver)
		ok <- true
	}()
	select {
	case okv := <-ok:
		if okv {
			//Reader:
			go func() {
				for {
					frame, err := LoadFrameFromStream(rv.in)
					if err != nil {
						log.Error("Invalid frame")
						log.Flush()
						os.Exit(1)
					}
					rv.olock.Lock()
					dest, ok := rv.seqnos[frame.SeqNo]
					rv.olock.Unlock()
					if ok {
						dest <- frame
					}
				}
			}()
			return rv, nil
		}
		return nil, errors.New("Bad router")
	case _ = <-time.After(5 * time.Second):
		log.Error("Timeout on router HELO")
		conn.Close()
		return nil, errors.New("Timeout on HELO")
	}
}

//ConnectOrExit is the same as Connect but will
//print an error message to stderr and exit if the connection
//fails
func ConnectOrExit(to string) *BW2Client {
	bw, err := Connect(to)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to local BW2 router:", err.Error())
		os.Exit(1)
	}
	return bw
}

//Sends a request frame and returns a frame that contains all the responses.
//Automatically closes the returned channel when there are no more responses.
func (cl *BW2Client) transact(req *Frame) chan *Frame {
	seqno := req.SeqNo
	inchan := make(chan *Frame, 3)
	outchan := make(chan *Frame, 3)
	cl.olock.Lock()
	cl.seqnos[seqno] = inchan
	req.WriteToStream(cl.out)
	cl.olock.Unlock()
	go func() {
		for {
			fr, ok := <-inchan
			if !ok {
				close(outchan)
				return
			}
			finished, ok := fr.GetFirstHeader("finished")
			if ok && finished == "true" {
				close(outchan)
				return
			}
			outchan <- fr
		}
	}()
	return outchan
}
func (cl *BW2Client) closeSeq(seqno int) {
	cl.olock.Lock()
	ch, ok := cl.seqnos[seqno]
	if ok {
		close(ch)
		delete(cl.seqnos, seqno)
	}
	cl.olock.Unlock()
}

func (cl *BW2Client) CreateEntity(p *CreateEntityParams) (string, []byte, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdMakeEntity, seqno)
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("contact", p.Contact)
	req.AddHeader("comment", p.Comment)
	for _, rvk := range p.Revokers {
		req.AddHeader("revoker", rvk)
	}
	if p.OmitCreationDate {
		req.AddHeader("omitcreationdate", "true")
	}
	rsp := cl.transact(req)
	fr, ok := <-rsp
	cl.closeSeq(seqno)
	if ok {
		if fr.Cmd == CmdResponse { //error
			msg, _ := fr.GetFirstHeader("reason")
			return "", nil, errors.New(msg)
		} else if len(fr.POs) != 1 {
			return "", nil, errors.New("bad response")
		}
		vk, _ := fr.GetFirstHeader("vk")
		po := fr.POs[0].PO

		return vk, po, nil
	}
	return "", nil, errors.New("reply channel closed")
}

func (cl *BW2Client) CreateDOT(p *CreateDOTParams) (string, *objects.DOT, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdMakeDot, seqno)
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("contact", p.Contact)
	req.AddHeader("comment", p.Comment)
	for _, rvk := range p.Revokers {
		req.AddHeader("revoker", rvk)
	}
	if p.OmitCreationDate {
		req.AddHeader("omitcreationdate", "true")
	}
	req.AddHeader("ttl", strconv.Itoa(int(p.TTL)))
	req.AddHeader("to", p.To)
	req.AddHeader("ispermission", strconv.FormatBool(p.IsPermission))
	if !p.IsPermission {
		req.AddHeader("uri", p.URI)
		req.AddHeader("accesspermissions", p.AccessPermissions)
	} else {
		panic("Not supported yet")
	}
	rsp := cl.transact(req)
	fr, ok := <-rsp
	cl.closeSeq(seqno)
	if ok {
		if fr.Cmd == CmdResponse { //error
			msg, _ := fr.GetFirstHeader("reason")
			return "", nil, errors.New(msg)
		} else if len(fr.ROs) != 1 {
			return "", nil, errors.New("bad response")
		}
		hash, _ := fr.GetFirstHeader("hash")
		ro := fr.ROs[0].RO

		return hash, ro.(*objects.DOT), nil
	}
	return "", nil, errors.New("reply channel closed")
}

func (cl *BW2Client) CreateDotChain(p *CreateDotChainParams) (string, *objects.DChain, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdMakeChain, seqno)
	req.AddHeader("ispermission", strconv.FormatBool(p.IsPermission))
	req.AddHeader("unelaborate", strconv.FormatBool(p.UnElaborate))
	for _, dot := range p.DOTs {
		req.AddHeader("dot", dot)
	}
	rsp := cl.transact(req)
	fr, ok := <-rsp
	cl.closeSeq(seqno)
	if ok {
		if fr.Cmd == CmdResponse { //error
			msg, _ := fr.GetFirstHeader("reason")
			return "", nil, errors.New(msg)
		} else if len(fr.ROs) != 1 {
			return "", nil, errors.New("bad response")
		}
		hash, _ := fr.GetFirstHeader("hash")
		ro := fr.ROs[0].RO

		return hash, ro.(*objects.DChain), nil
	}
	return "", nil, errors.New("reply channel closed")
}

func (cl *BW2Client) Publish(p *PublishParams) error {
	seqno := cl.GetSeqNo()
	cmd := CmdPublish
	if p.Persist {
		cmd = CmdPersist
	}
	req := CreateFrame(cmd, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}

	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	for _, po := range p.PayloadObjects {
		req.AddPayloadObject(po)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaborateFull
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	req.AddHeader("persist", strconv.FormatBool(p.Persist))
	rsp := cl.transact(req)
	fr, ok := <-rsp
	cl.closeSeq(seqno)
	if ok {
		status, _ := fr.GetFirstHeader("status")
		if status != "okay" {
			msg, _ := fr.GetFirstHeader("reason")
			return errors.New(msg)
		}
		return nil
	}
	return errors.New("receive channel closed")
}

func (cl *BW2Client) Subscribe(p *SubscribeParams) (chan *SimpleMessage, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdSubscribe, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}
	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaborateFull
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	if !p.LeavePacked {
		req.AddHeader("unpack", "true")
	}
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	rsp := cl.transact(req)
	//First response is the RESP frame
	fr, ok := <-rsp
	if ok {
		status, _ := fr.GetFirstHeader("status")
		if status != "okay" {
			msg, _ := fr.GetFirstHeader("reason")
			return nil, errors.New(msg)
		}
	} else {
		return nil, errors.New("receive channel closed")
	}
	//Generate converted output channel
	rv := make(chan *SimpleMessage, 10)
	go func() {
		for f := range rsp {
			sm := SimpleMessage{}
			sm.From, _ = f.GetFirstHeader("from")
			sm.URI, _ = f.GetFirstHeader("uri")
			sm.ROs = f.GetAllROs()
			poslice := make([]PayloadObject, f.NumPOs())
			errslice := make([]error, 0)
			for i := 0; i < f.NumPOs(); i++ {
				var err error
				poslice[i], err = f.GetPO(i)
				if err != nil {
					errslice = append(errslice, err)
				}
			}
			sm.POs = poslice
			sm.POErrors = errslice
			rv <- &sm
		}
		close(rv)
	}()
	return rv, nil
}

func (cl *BW2Client) SetEntity(keyfile []byte) (string, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdSetEntity, seqno)
	po := CreateBasePayloadObject(FromDotForm("1.0.1.2"), keyfile)
	req.AddPayloadObject(po)
	rsp := cl.transact(req)
	fr, ok := <-rsp
	cl.closeSeq(seqno)
	if ok {
		status, _ := fr.GetFirstHeader("status")
		if status != "okay" {
			msg, _ := fr.GetFirstHeader("reason")
			return "", errors.New(msg)
		}
		vk, _ := fr.GetFirstHeader("vk")
		return vk, nil
	}
	return "", errors.New("receive channel closed")
}

func (cl *BW2Client) SetEntityFileOrExit(filename string) string {
	rv, e := cl.SetEntityFile(filename)
	if e != nil {
		fmt.Fprintln(os.Stderr, "Could not set entity file:", e.Error())
		os.Exit(1)
	}
	return rv
}
func (cl *BW2Client) SetEntityFile(filename string) (string, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return cl.SetEntity(contents[1:])
}

func (cl *BW2Client) BuildChain(uri, permissions, to string) (chan *SimpleChain, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdBuildChain, seqno)
	req.AddHeader("uri", uri)
	req.AddHeader("to", to)
	req.AddHeader("accesspermissions", permissions)
	rv := make(chan *SimpleChain, 2)
	rsp := cl.transact(req)
	proc := func() {
		for fr := range rsp {
			hash, _ := fr.GetFirstHeader("hash")
			permissions, _ := fr.GetFirstHeader("permissions")
			to, _ := fr.GetFirstHeader("to")
			uri, _ := fr.GetFirstHeader("uri")
			sc := SimpleChain{
				Hash:        hash,
				Permissions: permissions,
				To:          to,
				URI:         uri,
			}
			rv <- &sc
		}
		cl.closeSeq(seqno)
		close(rv)
	}
	fr, ok := <-rsp
	if ok {
		status, _ := fr.GetFirstHeader("status")
		if status != "okay" {
			msg, _ := fr.GetFirstHeader("reason")
			return nil, errors.New(msg)
		}
		go proc()
		return rv, nil
	} else {
		return nil, errors.New("receive channel closed")
	}
}

func (cl *BW2Client) BuildAnyChainOrExit(uri, permissions, to string) *SimpleChain {
	rv, e := cl.BuildAnyChain(uri, permissions, to)
	if e != nil || rv == nil {
		fmt.Fprintf(os.Stderr, "Could not build chain to %s granting %s: %s", uri, permissions, e.Error())
		os.Exit(1)
	}
	return rv
}

func (cl *BW2Client) BuildAnyChain(uri, permissions, to string) (*SimpleChain, error) {
	rc, err := cl.BuildChain(uri, permissions, to)
	if err != nil {
		return nil, err
	}
	rv, ok := <-rc
	if ok {
		go func() {
			for _ = range rc {
			}
		}()
		return rv, nil
	}
	return nil, errors.New("No result")
}

func (cl *BW2Client) QueryOne(p *QueryParams) (*SimpleMessage, error) {
	rvc, err := cl.Query(p)
	if err != nil {
		return nil, err
	}
	v, ok := <-rvc
	if !ok {
		return nil, nil
	}
	go func() {
		for _ = range rvc {
		}
	}()
	return v, nil
}
func (cl *BW2Client) Query(p *QueryParams) (chan *SimpleMessage, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdQuery, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}
	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaborateFull
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	if !p.LeavePacked {
		req.AddHeader("unpack", "true")
	}
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	rsp := cl.transact(req)
	//First response is the RESP frame
	fr, ok := <-rsp
	if ok {
		status, _ := fr.GetFirstHeader("status")
		if status != "okay" {
			msg, _ := fr.GetFirstHeader("reason")
			return nil, errors.New(msg)
		}
	} else {
		return nil, errors.New("receive channel closed")
	}
	//Generate converted output channel
	rv := make(chan *SimpleMessage, 10)
	go func() {
		for f := range rsp {
			sm := SimpleMessage{}
			sm.From, _ = f.GetFirstHeader("from")
			sm.URI, _ = f.GetFirstHeader("uri")
			sm.ROs = f.GetAllROs()
			poslice := make([]PayloadObject, f.NumPOs())
			errslice := make([]error, 0)
			for i := 0; i < f.NumPOs(); i++ {
				var err error
				poslice[i], err = f.GetPO(i)
				if err != nil {
					errslice = append(errslice, err)
				}
			}
			sm.POs = poslice
			sm.POErrors = errslice
			rv <- &sm
		}
		close(rv)
	}()
	return rv, nil
}

func (cl *BW2Client) List(p *ListParams) (chan string, error) {
	seqno := cl.GetSeqNo()
	req := CreateFrame(CmdQuery, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}
	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaborateFull
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	rsp := cl.transact(req)
	//First response is the RESP frame
	fr, ok := <-rsp
	if ok {
		status, _ := fr.GetFirstHeader("status")
		if status != "okay" {
			msg, _ := fr.GetFirstHeader("reason")
			return nil, errors.New(msg)
		}
	} else {
		return nil, errors.New("receive channel closed")
	}
	//Generate converted output channel
	rv := make(chan string, 10)
	go func() {
		for f := range rsp {
			child, _ := f.GetFirstHeader("child")
			rv <- child
		}
		close(rv)
	}()
	return rv, nil
}

func FmtKey(key []byte) string {
	return base64.URLEncoding.EncodeToString(key)
}

func UnFmtKey(key string) ([]byte, error) {
	rv, err := base64.URLEncoding.DecodeString(key)
	if len(rv) != 32 {
		return nil, errors.New("Invalid length")
	}
	return rv, err
}

func FmtSig(sig []byte) string {
	return base64.URLEncoding.EncodeToString(sig)
}
func UnFmtSig(sig string) ([]byte, error) {
	rv, err := base64.URLEncoding.DecodeString(sig)
	if len(rv) != 64 {
		return nil, errors.New("Invalid length")
	}
	return rv, err
}

func FmtHash(hash []byte) string {
	return base64.URLEncoding.EncodeToString(hash)
}
func UnFmtHash(hash string) ([]byte, error) {
	rv, err := base64.URLEncoding.DecodeString(hash)
	if len(rv) != 32 {
		return nil, errors.New("Invalid length")
	}
	return rv, err
}
