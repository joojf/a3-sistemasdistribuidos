package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
)

const (
	HOST            = "localhost"
	PORT            = "12345"
	CONNECTION_TYPE = "tcp"
)

type Cliente struct {
	Id    string
	Nome  string
	Email string
	Role  string
}

type Request struct {
	Params string `json:"params"`
	Valor  []byte `json:"valor"`
}

type RequestCriarCliente struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type RequestEncerrarLeilao struct {
	Id string `json:"id"`
}

type RequestCriarLeilao struct {
	NomeVendedor    string `json:"nomeVendedor"`
	LanceMinimo     string `json:"lanceMinimo"`
	DescricaoArtigo string `json:"descricaoArtigo"`
}

type ResponseListarLeiloes struct {
	Leiloes []Leilao `json:"leiloes"`
}

type Leilao struct {
	Id              string
	Nome            string
	LanceMinimo     string
	DescricaoArtigo string
	Vencedor        string
	Status          string
}

type RequestCriarLance struct {
	IdLeilao string
	Email    string
	Valor    string
}

var leiloes []Leilao
var clientes []Cliente

func main() {
	fmt.Println("Iniciando servidor...")
	server, err := net.Listen(CONNECTION_TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal("Erro ao iniciar servidor: ", err)
	}

	defer server.Close()
	fmt.Println("Servidor iniciado com sucesso!")
	fmt.Println("Ouvindo conexões na porta: ", PORT)
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal("Erro ao aceitar conexão: ", err)
		}
		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Nova conexão recebida: ", conn.RemoteAddr())

	cliente := handleClient(conn)
	if cliente.Role == "vendedor" {
		go handleVendedor(conn, cliente)
	} else {
		go handleComprador(conn, cliente)
	}

}

func handleClient(conn net.Conn) Cliente {
	buffer := make([]byte, 1024)
	length, err := conn.Read(buffer)
	if err != nil {
		log.Fatal("Erro ao ler do cliente: ", err)
	}

	var cliente RequestCriarCliente
	json.Unmarshal(buffer[:length], &cliente)
	_, err = conn.Write([]byte("ok"))
	if err != nil {
		log.Fatal("Erro ao enviar para o cliente: ", err)
	}

	objetoCliente := Cliente{
		Id:    strconv.Itoa(rand.Intn(100)),
		Nome:  cliente.Nome,
		Email: cliente.Email,
		Role:  cliente.Role,
	}

	fmt.Println("Cliente adicionado: ", objetoCliente)

	clientes = append(clientes, objetoCliente)
	return objetoCliente
}

func handleVendedor(conn net.Conn, cliente Cliente) {
	for {
		message := handleMessage(conn)
		var request Request
		json.Unmarshal([]byte(message), &request)

		switch request.Params {
		case "criarLeilao":
			var requestCriarLeilao RequestCriarLeilao
			json.Unmarshal(request.Valor, &requestCriarLeilao)
			itemLeilao := Leilao{
				Id:              strconv.Itoa(rand.Intn(100)),
				Nome:            requestCriarLeilao.NomeVendedor,
				LanceMinimo:     requestCriarLeilao.LanceMinimo,
				DescricaoArtigo: requestCriarLeilao.DescricaoArtigo,
				Vencedor:        "",
				Status:          "aberto",
			}
			leiloes = append(leiloes, itemLeilao)
			fmt.Println("Leilão criado: ", itemLeilao)
			conn.Write([]byte("Leilão criado com sucesso!"))
		case "encerrarLeilao":
			var requestEncerrarLeilao RequestEncerrarLeilao
			json.Unmarshal(request.Valor, &requestEncerrarLeilao)
			for i, leilao := range leiloes {
				if leilao.Id == requestEncerrarLeilao.Id {
					leiloes[i].Status = "encerrado"
					leiloes = append(leiloes[:i], leiloes[i+1:]...)
					conn.Write([]byte("Leilão encerrado com sucesso!"))
					break
				}
			}
		case "sair":
			fmt.Println("Cliente desconectado: ", cliente)
			conn.Write([]byte("Cliente desconectado!"))
			return
		default:
			fmt.Println("Mensagem desconhecida: ", request.Params)
			conn.Write([]byte("Mensagem desconhecida!"))
		}
	}
}

func handleComprador(conn net.Conn, cliente Cliente) {
	for {
		message := handleMessage(conn)
		var request Request
		json.Unmarshal([]byte(message), &request)

		switch request.Params {
		case "listarLeiloes":
			var response ResponseListarLeiloes
			for _, leilao := range leiloes {
				if leilao.Status == "aberto" {
					response.Leiloes = append(response.Leiloes, leilao)
				}
			}
			json, _ := json.Marshal(response)
			conn.Write(json)
		case "criarLance":
			var requestCriarLance RequestCriarLance
			json.Unmarshal(request.Valor, &requestCriarLance)
			for _, leilao := range leiloes {
				if leilao.Id == requestCriarLance.IdLeilao {
					if leilao.Status == "aberto" {
						if requestCriarLance.Valor > leilao.LanceMinimo {
							leilao.LanceMinimo = requestCriarLance.Valor
							leilao.Vencedor = cliente.Email
							conn.Write([]byte("Lance criado com sucesso!"))
							break
						} else {
							conn.Write([]byte("Valor do lance menor que o lance minimo!"))
							break
						}
					}
				}
			}
		case "sair":
			fmt.Println("Cliente desconectado: ", clientes[0])
			conn.Write([]byte("Cliente desconectado!"))
			return
		default:
			fmt.Println("Mensagem desconhecida: ", request.Params)
			conn.Write([]byte("Mensagem desconhecida!"))
		}

	}
}

func handleMessage(conn net.Conn) string {
	buffer := make([]byte, 1024)
	length, err := conn.Read(buffer)
	if err != nil {
		log.Fatal("Erro ao ler do cliente: ", err)
	}
	return string(buffer[:length])
}
