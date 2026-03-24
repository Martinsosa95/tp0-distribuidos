
import logging

from common.utils import SUCCESS_ACK


class Protocolo:
    BATCH_APUESTAS=1
    COD_EXITO=1
    NOTIFICACION=2
    CONSULTA=3


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
    
    def recibir_mensaje(self):
        logging.debug("action: recibir_apuesta_lectura | result: in_progress")
        data = self.recv_exactly(4)
        if data is None:
            return None, None

        payload_length = int.from_bytes(data, byteorder='big')
        payload = self.recv_exactly(payload_length)
        if payload is None:
            return None, None
        
        opcode = payload[0]
        cuerpo = payload[1:]

        if opcode == self.BATCH_APUESTAS:
            batch = cuerpo.decode().split('\n')
            return opcode, [self.parse_bet_line(line) for line in batch if line.strip()]

        elif opcode == self.NOTIFICACION or opcode == self.CONSULTA:
            return opcode, cuerpo.decode('utf-8')
    
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

    def enviar_ack(self, codigo):
        """ Envia un mensaje de ACK al cliente:
        0x00 OK
        0x01 ERROR
        0x02 SORTEO NO LISTO"""
        self.socket.sendall(bytes([codigo]))

    def enviar_ganadores(self, ganadores):
        self.enviar_ack(SUCCESS_ACK)
        payload = '\n'.join(ganadores).encode('utf-8')
        payload_length = len(payload)
        self.socket.sendall(payload_length.to_bytes(4, byteorder='big') + payload)
