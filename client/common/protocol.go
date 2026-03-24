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

	log.Infof(
		"action: send_data | result: in_progress | payload: %s",
		payloadStr,
	)

	payload := []byte(payloadStr)

	l := uint32(len(payload))
	header := []byte{
		byte(l >> 24),
		byte(l >> 16),
		byte(l >> 8),
		byte(l),
	}
	log.Infof(
		"action: send_data_header | result: in_progress ",
	)
	if _, err := p.conn.Write(header); err != nil {
		return err
	}
	log.Infof(
		"action: send_data_payload | result: in_progress ",
	)
	if _, err := p.conn.Write(payload); err != nil {
		return err
	}

	log.Infof(
		"Esperando ACK del servidor... | action: wait_for_ack | result: in_progress ",
	)
	ack := make([]byte, 1)
	if _, err := io.ReadFull(p.conn, ack); err != nil {
		return err
	}
	log.Infof(
		"ACK recibido del servidor. | action: wait_for_ack | result: success ",
	)
	return nil
}

func (p *Protocolo) Close() error {
	return p.conn.Close()
}
