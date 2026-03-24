
import logging


class Protocolo:
    def __init__(self, socket):
        self.socket = socket

    def recv_exactly(self, num_bytes):
        data = bytearray()
        while len(data) < num_bytes:
            packet = self.socket.recv(num_bytes - len(data))
            if not packet:
                return None
            data.extend(packet)
        return data

    def recibir_apuesta(self):
        logging.debug("action: recibir_apuesta_lectura | result: in_progress")
        data = self.recv_exactly(4)
        if data is None:
            return None

        payload_length = int.from_bytes(data, byteorder='big')
        payload = self.recv_exactly(payload_length)
        if payload is None:
            return None
        return self.parse_bet_line(payload.decode())
    
    def recibir_batch(self):
        logging.debug("action: recibir_apuesta_lectura | result: in_progress")
        data = self.recv_exactly(4)
        if data is None:
            return None

        payload_length = int.from_bytes(data, byteorder='big')
        payload = self.recv_exactly(payload_length)
        if payload is None:
            return None
        
        batch = payload.decode().split('\n')
        return [self.parse_bet_line(line) for line in batch if line.strip()]
    
    def parse_bet_line(self, line):
        fields = line.split('|')
        return {
            'agency': fields[0],
            'first_name': fields[1],
            'last_name': fields[2],
            'document': fields[3],
            'birthdate': fields[4],
            'number': fields[5],
        }

    def enviar_ack(self):
        self.socket.sendall(b'\x00')
