import socket
print("Hello World!")
hostname = socket.gethostname()
ip_address = socket.gethostbyname(hostname)
print(ip_address)