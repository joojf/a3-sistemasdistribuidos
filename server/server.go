package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
)

const (
	HOST            = "localhost"
	PORT            = "12345"
	CONNECTION_TYPE = "tcp"
)

type Cliente struct {
	Id    int
	Nome  string
	Email string
	Role  string
}

type Request struct {
	Params string `json:"params"`
	Valor  int    `json:"valor"`
}

type RequestCriarCliente struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type RequestEncerrarLeilao struct {
	Id int `json:"id"`
}

type RequestCriarLeilao struct {
	NomeVendedor    string `json:"nomeVendedor"`
	LanceMinimo     int    `json:"lanceMinimo"`
	DescricaoArtigo string `json:"descricaoArtigo"`
}

type ResponseListarLeiloes struct {
	Leiloes []Leilao `json:"leiloes"`
}

type Leilao struct {
	Id              int
	Nome            string
	LanceMinimo     int
	DescricaoArtigo string
	Vencedor        string
}

type Lance struct {
	IdLeilao int
	Email    string
	Valor    float64
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
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal("Erro ao ler do cliente: ", err)
	}

	var cliente RequestCriarCliente
	err = json.Unmarshal(buffer, &cliente)
	if err != nil {
		log.Fatal("Erro ao converter json: ", err)
	}

	novoCliente := Cliente{
		Id:    rand.Intn(100),
		Nome:  cliente.Nome,
		Email: cliente.Email,
		Role:  cliente.Role,
	}
	clientes = append(clientes, novoCliente)
	fmt.Println("Cliente criado: ", novoCliente)
	return novoCliente
}

func handleVendedor(conn net.Conn, cliente Cliente) {
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			log.Fatal("Erro ao ler do cliente: ", err)
		}

		var request Request
		err = json.Unmarshal(buffer, &request)
		if err != nil {
			log.Fatal("Erro ao converter json: ", err)
		}
		switch request.Params {
		case "criarLeilao":
			var request RequestCriarLeilao
			handleCriarLeilao(conn, cliente, request.LanceMinimo, request.DescricaoArtigo)
		case "encerrarLeilao":
			var request RequestEncerrarLeilao
			handleEncerrarLeilao(conn, cliente, request.Id)
		case "sair":
			fmt.Println("Cliente saiu: ", cliente)
			return
		default:
			fmt.Println("Parâmetro inválido: ", request.Params)
		}
	}
}

func handleComprador(conn net.Conn, cliente Cliente) {
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			log.Fatal("Erro ao ler do cliente: ", err)
		}

		var request Request
		err = json.Unmarshal(buffer, &request)
		if err != nil {
			log.Fatal("Erro ao converter json: ", err)
		}

		switch request.Params {
		case "listarLeiloes":
			var response ResponseListarLeiloes
			handleListarLeiloes(conn, cliente, &response)
		case "fazerLance":
			var lance Lance
			handleFazerLance(conn, cliente, request.Valor, lance)
		case "sair":
			fmt.Println("Cliente saiu: ", cliente)
			return
		default:
			fmt.Println("Parâmetro inválido: ", request.Params)
		}
	}
}

func handleCriarLeilao(conn net.Conn, cliente Cliente, lanceMinimo int, descricaoArtigo string) {
	leilao := Leilao{
		Id:              rand.Intn(100),
		Nome:            cliente.Nome,
		LanceMinimo:     lanceMinimo,
		DescricaoArtigo: descricaoArtigo,
		Vencedor:        "",
	}
	leiloes = append(leiloes, leilao)
	fmt.Println("Leilão criado: ", leilao)
	conn.Write([]byte("Leilão criado com sucesso!"))
}

func handleEncerrarLeilao(conn net.Conn, cliente Cliente, id int) {
	for _, leilao := range leiloes {
		if leilao.Id == id {
			leilao.Vencedor = cliente.Nome
			fmt.Println("Leilão encerrado: ", leilao)
			conn.Write([]byte("Leilão encerrado com sucesso!"))
			return
		}
	}
	conn.Write([]byte("Leilão não encontrado!"))
}

func handleListarLeiloes(conn net.Conn, cliente Cliente, response *ResponseListarLeiloes) {
	for _, leilao := range leiloes {
		response.Leiloes = append(response.Leiloes, leilao)
	}
	buffer, err := json.Marshal(response)
	if err != nil {
		log.Fatal("Erro ao converter json: ", err)
	}
	conn.Write(buffer)
}

func handleFazerLance(conn net.Conn, cliente Cliente, valor int, lance Lance) {
	for _, leilao := range leiloes {
		if leilao.Id == lance.IdLeilao {
			if valor >= leilao.LanceMinimo {
				leilao.Vencedor = lance.Email
				fmt.Println("Lance feito: ", lance)
				conn.Write([]byte("Lance feito com sucesso!"))
				return
			} else {
				conn.Write([]byte("Valor do lance menor que o mínimo!"))
				return
			}
		}
	}
	conn.Write([]byte("Leilão não encontrado!"))
}
