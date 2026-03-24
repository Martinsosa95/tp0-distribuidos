package common

import (
	"fmt"
	"io"
	"net"
	"strings"
)

type Apuesta struct {
	Agencia   string
	Nombre    string
	Apellido  string
	Documento string
	Nacimiento string
	Numero    string
}

type Protocolo struct {
	conn net.Conn
}


const (
	OpcodeBatch       byte = 0x01
	OpcodeFinEnvio    byte = 0x02
	OpcodeConsulta    byte = 0x03
)

const (
	AckOK             byte = 0x00
	AckError          byte = 0x01
	AckNotReady		  byte = 0x02
)

var ErrNotReady = fmt.Errorf("Sorteo no listo")

func Connect(address string) (*Protocolo, error) {
	conn, err := net.Dial("tcp", address)
	if err == nil {
			return &Protocolo{conn: conn}, nil
	}else  {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}
	return &Protocolo{conn: conn}, nil
}

func (p *Protocolo) EnviarApuesta(apuestas []Apuesta) error {
	var payloadLines []string

	for _, apuesta := range apuestas {
		line := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
			apuesta.Agencia,
			apuesta.Nombre,
			apuesta.Apellido,
			apuesta.Documento,
			apuesta.Nacimiento,
			apuesta.Numero,
		)
		payloadLines = append(payloadLines, line)
	}
	
	payloadStr := strings.Join(payloadLines, "\n")
	payload := []byte(payloadStr)

	l := uint32(len(payload))
	header := []byte{
		byte(l >> 24),
		byte(l >> 16),
		byte(l >> 8),
		byte(l),
	}

	if _, err := p.conn.Write(header); err != nil {
		return err
	}

	if _, err := p.conn.Write(payload); err != nil {
		return err
	}

	ack := make([]byte, 1)
	io.ReadFull(p.conn, ack)
	if ack[0] != AckOK {
		return fmt.Errorf("Error from server: %v", ack[0])
	}
	return nil
}

func (p *Protocolo) EnviarNotificacion(agencia string) ([]string, error) {
	payload := append([]byte{OpcodeFinEnvio}, []byte(agencia)...)
	l := uint32(len(payload))
	header := []byte{
		byte(l >> 24),
		byte(l >> 16),
		byte(l >> 8),
		byte(l),
	}

	if _, err := p.conn.Write(header); err != nil {
		return nil, err
	}

	if _, err := p.conn.Write(payload); err != nil {
		return nil, err
	}

	ack := make([]byte, 1)
	io.ReadFull(p.conn, ack)
	if ack[0] != AckOK {
		return nil, fmt.Errorf("Error from server: %v", ack[0])
	}
	return nil, nil
}

func (p *Protocolo) EnviarConsulta(agencia string) ([]string, error) {
	payload := append([]byte{OpcodeConsulta}, []byte(agencia)...)
	l := uint32(len(payload))
	header := []byte{
		byte(l >> 24),
		byte(l >> 16),
		byte(l >> 8),
		byte(l),
	}

	if _, err := p.conn.Write(header); err != nil {
		return nil, err
	}

	if _, err := p.conn.Write(payload); err != nil {
		return nil, err
	}

	ack := make([]byte, 1)
	io.ReadFull(p.conn, ack)
	if ack[0] == AckNotReady {
		return nil, ErrNotReady 
	} else if ack[0] != AckOK {
		return nil, fmt.Errorf("Error from server: %v", ack[0])
	}

	respHeader := make([]byte, 4)
	if _, err := io.ReadFull(p.conn, respHeader); err != nil {
		return nil, err
	}

	respLen := uint32(respHeader[0])<<24 | uint32(respHeader[1])<<16 | uint32(respHeader[2])<<8 | uint32(respHeader[3])
	if respLen == 0 {
		return []string{}, nil
	}

	respPayload := make([]byte, respLen)
	io.ReadFull(p.conn, respPayload)
	return strings.Split(string(respPayload), ","), nil
}

func (p *Protocolo) Close() error {
	return p.conn.Close()
}
