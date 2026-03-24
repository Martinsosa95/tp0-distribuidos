package common

import (
	"io"
	"os"
	"encoding/csv"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchMaxAmount int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {

	file, err := os.Open("/data/agency.csv")
	if err != nil {
		log.Errorf("action: abrir_archivo | result: fail | error: %v", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var batch []Apuesta
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("action: leer_archivo | result: fail | error: %v", err)
			return
		}


		apuesta := Apuesta{
			Agencia:    c.config.ID,
			Nombre:     record[0],
			Apellido:   record[1],
			Documento:  record[2],
			Nacimiento: record[3],
			Numero:     record[4],
		}
		batch = append(batch, apuesta)
		
		if len(batch) >= c.config.BatchMaxAmount {
			c.enviarBatch(batch)
			batch = []Apuesta{}

			time.Sleep(c.config.LoopPeriod)
		}

	}
	if len(batch) > 0 {
		c.enviarBatch(batch)
	}

	c.enviarNotificacion(c.config.ID)

	for{

		dni, err := c.recibirGanadores()

		if err == ErrNotReady {
			log.Infof("action: recibir_ganadores | result: not_ready | client_id: %v", c.config.ID)
			time.Sleep(c.config.LoopPeriod)
			continue
		} else if err != nil {
			log.Errorf("action: recibir_ganadores | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(dni))
		break
	}
}

func (c *Client) enviarBatch(batch []Apuesta) {
	protocolo, err := Connect(c.config.ServerAddress)
	if err != nil {
		log.Errorf("action: conectar_servidor | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer protocolo.Close()

	if err := protocolo.EnviarApuesta(batch); err != nil {
		log.Errorf("action: enviar_apuesta | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	log.Infof("action: enviar_apuesta | result: success | client_id: %v | batch_size: %v", c.config.ID, len(batch))
}

func (c *Client) enviarNotificacion(agencia string) {
	protocolo, err := Connect(c.config.ServerAddress)
	if err != nil {
		log.Errorf("action: conectar_servidor | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer protocolo.Close()
	protocolo.EnviarNotificacion(agencia)
}

func (c *Client) recibirGanadores() ([]string, error) {
	protocolo, err := Connect(c.config.ServerAddress)
	if err != nil {
		log.Errorf("action: conectar_servidor | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return nil, err
	}
	defer protocolo.Close()
	return protocolo.EnviarConsulta(c.config.ID)
}
