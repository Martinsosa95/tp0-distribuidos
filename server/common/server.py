import socket
import logging
import signal
from common.protocol import Protocolo
from common.utils import Bet, store_bets, SUCCESS_ACK, ERROR_ACK, NOT_READY_ACK, load_bets, has_won


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        signal.signal(signal.SIGTERM, self.__handle_signal)

        self.agencies = set()
        self.agencies_finished = set()
        self.sorteo = False

    def __handle_signal(self, signum, frame):
        """Handle SIGTERM signal to gracefully shutdown the server"""
        logging.info("action: signal_handler | result: success | signal: SIGTERM")
        self._running = False
        try:
            self._server_socket.shutdown(socket.SHUT_RDWR)
            self._server_socket.close()
            logging.info("action: close_resource | result: success | resource: server_socket")
        except Exception as e:
            logging.error(f"action: close_resource | result: fail | resource: server_socket | error: {e}")

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while self._running:
            client_sock = self.__accept_new_connection()
            if client_sock:
                self.__handle_client_connection(client_sock)
        logging.info("action: server_shutdown | result: success")

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            protocolo = Protocolo(client_sock)
            logging.info("action: handle_connection | result: in_progress")
            opcode, data_apuestas = protocolo.recibir_mensaje()

            if opcode == Protocolo.BATCH_APUESTAS and data_apuestas:
                try:
                    bets = [Bet(**data_apuesta) for data_apuesta in data_apuestas]

                    for bet in bets:
                        self.agencies.add(bet.agency)
                    
                    store_bets(bets)
                    protocolo.enviar_ack(SUCCESS_ACK)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                except Exception as e:
                    logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(data_apuestas) if data_apuestas else 0}")
                    protocolo.enviar_ack(ERROR_ACK)
            elif opcode == Protocolo.NOTIFICACION and data_apuestas:
                self.agencies_finished += 1
                if self.agencies_finished == len(self.agencies):
                    self.sorteo = True
                    logging.info("action: sorteo | result: success")
                protocolo.enviar_ack(SUCCESS_ACK)
            elif opcode == Protocolo.CONSULTA and data_apuestas:
                if self.sorteo:
                    agencia_id = data_apuestas
                    ganadores = []
                    for bet in load_bets():
                        if bet.agency == int(agencia_id) and has_won(bet):
                            ganadores.append(bet.document)
                    protocolo.enviar_ganadores(ganadores)
                else:
                    protocolo.enviar_ack(NOT_READY_ACK)
                    logging.info("action: consulta_recibida | result: success | sorteo_listo: false")



        except Exception as e:
            logging.error(f"action: handle_connection | result: fail | error: {e}")
        finally:
            client_sock.close()
            logging.info("action: close_resource | result: success | resource: client_socket")

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except OSError as e:
            if not self._running:
                return None
            else:
                logging.error(f'action: accept_connections | result: fail | error: {e}')
                raise e
