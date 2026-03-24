package common

import (
	"fmt"
	"io"
	"net"
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

func Connect(address string) (*Protocolo, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}
	return &Protocolo{conn: conn}, nil
}

func (p *Protocolo) EnviarApuesta(apuesta Apuesta) error {

	payloadStr := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		apuesta.Agencia,
		apuesta.Nombre,
		apuesta.Apellido,
		apuesta.Documento,
		apuesta.Nacimiento,
		apuesta.Numero,
	)

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
	if _, err := io.ReadFull(p.conn, ack); err != nil {
		return err
	}
	return nil
}

func (p *Protocolo) Close() error {
	return p.conn.Close()
}
