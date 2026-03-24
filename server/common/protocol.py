
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
        data = self.recv_exactly(4)
        if data is None:
            return None

        payload_length = int.from_bytes(data, byteorder='big')
        payload = self.recv_exactly(payload_length)
        if payload is None:
            return None

        fields = payload.decode().split('|')
        return {
            'agencia': fields[0],
            'nombre': fields[1],
            'apellido': fields[2],
            'dni': fields[3],
            'nacimiento': fields[4],
            'numero': fields[5],
        }

    def enviar_ack(self):
        self.socket.sendall(b'ACK')
