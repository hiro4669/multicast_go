package mcast

const (
	SND = 0
	REQ = 1
	ACK = 2
)

type Packet struct {
	ptype  uint8
	sender uint8
	rtime  float32
	reqId  uint8
}

func NewPacket(ptype, sender uint8, rtime float32, reqId uint8) *Packet {
	return &Packet{ptype: ptype, sender: sender, rtime: rtime, reqId: reqId}
}

func (p *Packet) GetType() uint8 {
	return p.ptype
}
func (p *Packet) GetSender() uint8 {
	return p.sender
}

func (p *Packet) GetTime() float32 {
	return p.rtime
}

func (p *Packet) GetReqId() uint8 {
	return p.reqId
}
