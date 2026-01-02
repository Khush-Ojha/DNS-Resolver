package dns

import (
	"bytes"
	"encoding/binary"
	"strings"
)

type Header struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type Question struct {
	Name  string
	Type  uint16
	Class uint16
}

type ResourceRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	RData []byte
}

// ToName converts the RData bytes into a readable domain string (used for CNAME/NS records)
func (r *ResourceRecord) ToName() string {
	// Basic parser for the RData.
	// In a production system, this needs the full original packet for decompression pointers.
	// Here we implement a simplified reader for standalone names.
	var parts []string
	reader := bytes.NewReader(r.RData)
	for {
		b, err := reader.ReadByte()
		if err != nil || b == 0 {
			break
		}
		length := int(b)
		buf := make([]byte, length)
		reader.Read(buf)
		parts = append(parts, string(buf))
	}
	return strings.Join(parts, ".")
}

type Message struct {
	Header      Header
	Questions   []Question
	Answers     []ResourceRecord
	Authorities []ResourceRecord
	Additionals []ResourceRecord
}

func NewQuery(domain string) *Message {
	return &Message{
		Header: Header{
			ID:      1234,
			Flags:   0,
			QDCount: 1,
		},
		Questions: []Question{
			{
				Name:  domain,
				Type:  1, // Type A (IPv4)
				Class: 1, // Class IN
			},
		},
	}
}

func (m *Message) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, m.Header); err != nil {
		return nil, err
	}
	for _, q := range m.Questions {
		encodedName := encodeDNSName(q.Name)
		buf.Write(encodedName)
		binary.Write(buf, binary.BigEndian, q.Type)
		binary.Write(buf, binary.BigEndian, q.Class)
	}
	return buf.Bytes(), nil
}

func encodeDNSName(domain string) []byte {
	parts := strings.Split(domain, ".")
	var buf bytes.Buffer
	for _, part := range parts {
		buf.WriteByte(byte(len(part)))
		buf.WriteString(part)
	}
	buf.WriteByte(0)
	return buf.Bytes()
}

func FromBytes(data []byte) (*Message, error) {
	m := &Message{}
	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.BigEndian, &m.Header); err != nil {
		return nil, err
	}
	for i := 0; i < int(m.Header.QDCount); i++ {
		q, err := parseQuestion(reader, data)
		if err != nil {
			return nil, err
		}
		m.Questions = append(m.Questions, q)
	}
	for i := 0; i < int(m.Header.ANCount); i++ {
		r, err := parseRecord(reader, data)
		if err != nil {
			return nil, err
		}
		m.Answers = append(m.Answers, r)
	}
	for i := 0; i < int(m.Header.NSCount); i++ {
		r, err := parseRecord(reader, data)
		if err != nil {
			return nil, err
		}
		m.Authorities = append(m.Authorities, r)
	}
	for i := 0; i < int(m.Header.ARCount); i++ {
		r, err := parseRecord(reader, data)
		if err != nil {
			return nil, err
		}
		m.Additionals = append(m.Additionals, r)
	}
	return m, nil
}

func parseQuestion(reader *bytes.Reader, fullData []byte) (Question, error) {
	name, err := decodeName(reader, fullData)
	if err != nil {
		return Question{}, err
	}
	var type_, class uint16
	binary.Read(reader, binary.BigEndian, &type_)
	binary.Read(reader, binary.BigEndian, &class)
	return Question{Name: name, Type: type_, Class: class}, nil
}

func parseRecord(reader *bytes.Reader, fullData []byte) (ResourceRecord, error) {
	name, err := decodeName(reader, fullData)
	if err != nil {
		return ResourceRecord{}, err
	}
	var type_, class uint16
	var ttl uint32
	var dataLen uint16
	binary.Read(reader, binary.BigEndian, &type_)
	binary.Read(reader, binary.BigEndian, &class)
	binary.Read(reader, binary.BigEndian, &ttl)
	binary.Read(reader, binary.BigEndian, &dataLen)

	// IMPORTANT: We read the data immediately into a buffer
	data := make([]byte, dataLen)
	reader.Read(data)

	return ResourceRecord{Name: name, Type: type_, Class: class, TTL: ttl, RData: data}, nil
}

func decodeName(reader *bytes.Reader, fullData []byte) (string, error) {
	var parts []string
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0 {
			break
		}
		if b&0xC0 == 0xC0 {
			b2, _ := reader.ReadByte()
			offset := ((uint16(b) ^ 0xC0) << 8) | uint16(b2)
			newReader := bytes.NewReader(fullData)
			newReader.Seek(int64(offset), 0)
			name, _ := decodeName(newReader, fullData)
			parts = append(parts, name)
			return strings.Join(parts, "."), nil
		}
		length := int(b)
		buf := make([]byte, length)
		reader.Read(buf)
		parts = append(parts, string(buf))
	}
	return strings.Join(parts, "."), nil
}
