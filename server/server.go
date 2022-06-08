package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strconv"
)

const (
	HOST            = "localhost"
	PORT            = "12345"
	CONNECTION_TYPE = "tcp"
)

const (
	LEILAO_ATIVO     = "ATIVO"
	LEILAO_ENCERRADO = "ENCERRADO"
)

type DbCliente struct {
	Nome  string
	Email string
	Role  string
	Id    string
}

type Message struct {
	Operacao string `json:"operacao"`
	Message  []byte `json:"message"`
}
type MessageCriarCliente struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type MessageEncerrarLeilao struct {
	Id string `json:"id"`
}

type MessageCriarLeilao struct {
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Valor     string `json:"valor"`
}

type MessageRespostaListarLeiloesVendedor struct {
	Leiloes []ItemLeilaoVendedor `json:"leiloes"`
}

type MessageRespostaListarLeiloesComprador struct {
	Leiloes []ItemLeilaoComprador `json:"leiloes"`
}

type ItemLeilaoVendedor struct {
	Id   string
	Nome string
}

type ItemLeilaoComprador struct {
	Id           string
	Nome         string
	Descricao    string
	ValorInicial int
	LanceAtual   Lance
}

type ItemLeilaoDB struct {
	Id           string
	IdVendedor   string
	Nome         string
	Descricao    string
	LanceAtual   Lance
	ValorInicial int
	Status       string
}

type Lance struct {
	Email string
	Valor int
}

type MessageDarLance struct {
	Id    string `json:"id"`
	Valor int    `json:"valor"`
}

var itensLeilaoDB []ItemLeilaoDB
var clientes []DbCliente

func main() {
	fmt.Println("Iniciando servidor")
	server, err := net.Listen(CONNECTION_TYPE, HOST+":"+PORT)
	handleError(err, "Erro ao escutar host:")

	defer server.Close()
	fmt.Println("Escutando em " + server.Addr().String())
	fmt.Println("Esperando conexões...")
	for {
		connection, err := server.Accept()

		if err != nil {
			handleError(err, "Erro ao aceitar conexão:")
			connection.Close()
			continue
		}

		fmt.Println("Conexão aceita do cliente: ", connection.RemoteAddr())
		go processClient(connection)
	}
}
func processClient(connection net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Ocorreu um erro com a conexão do cliente:", connection.RemoteAddr(), "o erro foi:", err)
		}
	}()
	cliente := handleAuthentication(connection)
	if cliente.Role == "vendedor" {
		handleVendedor(connection, cliente)
	} else {
		handleComprador(connection, cliente)
	}
}

func handleError(err error, message string) {
	if err != nil {
		fmt.Println(message, err.Error())
	}
}

func handleConnectionError(connection net.Conn, err error, message string) {
	if err != nil {
		handleError(err, message)
		connection.Close()
		panic(err.Error())
	}
}

func handleAuthentication(connection net.Conn) DbCliente {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	fmt.Println("Cliente conectado: ", string(buffer[:mLen]))
	handleConnectionError(connection, err, "Perdemos a conexão com o cliente")

	var cliente MessageCriarCliente
	json.Unmarshal(buffer[:mLen], &cliente)
	_, err = connection.Write([]byte(cliente.Nome + " você já está conectado e já pode fazer leilões"))
	handleConnectionError(connection, err, "Error writing")
	dbCliente, exists := clienteExiste(cliente)
	if exists {
		fmt.Println("Cliente já existe")
		return dbCliente.(DbCliente)
	}
	fmt.Println("Novo cliente adicionado: ", cliente)

	novoCliente := &DbCliente{
		Nome:  cliente.Nome,
		Email: cliente.Email,
		Role:  cliente.Role,
		Id:    generateRandomId(),
	}
	clientes = append(clientes, *novoCliente)
	return *novoCliente
}

func clienteExiste(cliente MessageCriarCliente) (interface{}, bool) {
	for _, value := range clientes {

		if value.Email == cliente.Email && value.Nome == cliente.Nome && value.Role == cliente.Role {
			return value, true
		}
	}
	return nil, false
}

func handleComprador(connection net.Conn, comprador DbCliente) {
	for {
		message := handleSocketMessage(connection)
		var jsonMsg Message
		json.Unmarshal([]byte(message), &jsonMsg)
		switch jsonMsg.Operacao {
		case "LISTAR_LEILOES":
			var leiloesAEnviar []ItemLeilaoComprador
			for i, ild := range itensLeilaoDB {
				if ild.Status == LEILAO_ATIVO {
					leiloesAEnviar = append(leiloesAEnviar, ItemLeilaoComprador{
						Id:           itensLeilaoDB[i].Id,
						Nome:         itensLeilaoDB[i].Nome,
						Descricao:    itensLeilaoDB[i].Descricao,
						LanceAtual:   itensLeilaoDB[i].LanceAtual,
						ValorInicial: itensLeilaoDB[i].ValorInicial,
					})
				}
			}

			message, _ := json.Marshal(&MessageRespostaListarLeiloesComprador{
				Leiloes: leiloesAEnviar,
			})
			sendMessageToClient(connection, string(message))
		case "DAR_LANCE":
			var lanceLeilao MessageDarLance
			json.Unmarshal(jsonMsg.Message, &lanceLeilao)
			fmt.Println("Leilão recebido: ", lanceLeilao)
			var ehMaiorLance bool
			maiorQueValorInicial := true
			for i, value := range itensLeilaoDB {
				if value.Id == lanceLeilao.Id {
					fmt.Println("Leilão encontrado", value)

					apostaAtual := itensLeilaoDB[i].LanceAtual
					fmt.Println("Aposta atual: ", apostaAtual.Valor)

					if lanceLeilao.Valor < itensLeilaoDB[i].ValorInicial {
						maiorQueValorInicial = false
						break
					}
					if lanceLeilao.Valor > apostaAtual.Valor {
						itensLeilaoDB[i].LanceAtual.Email = comprador.Email
						itensLeilaoDB[i].LanceAtual.Valor = lanceLeilao.Valor
						ehMaiorLance = true
						fmt.Println("Novo lance: ", itensLeilaoDB[i].LanceAtual)
					} else {
						ehMaiorLance = false
					}
					break
				}
			}
			var message string

			if ehMaiorLance {
				message = "O seu lance foi o maior até o momento"
			} else {
				message = "O seu lance foi aceito"
			}
			if !maiorQueValorInicial {
				message = "O seu lance não foi aceito, pois ele é menor que o valor inicial"
			}

			sendMessageToClient(connection, message)
		}
	}
}

func handleVendedor(connection net.Conn, vendedor DbCliente) {
	for {
		message := handleSocketMessage(connection)
		var jsonMsg Message
		json.Unmarshal([]byte(message), &jsonMsg)
		switch jsonMsg.Operacao {
		case "CRIAR_LEILAO":
			var itemLeilao MessageCriarLeilao
			json.Unmarshal(jsonMsg.Message, &itemLeilao)
			valorInicial, err := strconv.Atoi(itemLeilao.Valor)
			if err != nil {
				valorInicial = 0
				fmt.Println("Erro ao converter valor inicial", itemLeilao)
			}
			itemLeilaoDB := ItemLeilaoDB{
				Id:           generateRandomId(),
				Nome:         itemLeilao.Nome,
				Descricao:    itemLeilao.Descricao,
				IdVendedor:   vendedor.Id,
				ValorInicial: valorInicial,
				Status:       LEILAO_ATIVO,
			}
			itensLeilaoDB = append(itensLeilaoDB, itemLeilaoDB)
			message := "O leilão com nome " + itemLeilao.Nome + " foi criado com sucesso"
			sendMessageToClient(connection, message)

		case "ENCERRAR_LEILAO":
			var idLeilao MessageEncerrarLeilao
			json.Unmarshal(jsonMsg.Message, &idLeilao)
			var exists = false
			var foundItem ItemLeilaoDB
			for i, value := range itensLeilaoDB {
				if value.Id == idLeilao.Id && value.IdVendedor == vendedor.Id {
					exists = true
					itensLeilaoDB[i].Status = LEILAO_ENCERRADO
					foundItem = value
				}
			}

			if !exists {
				message := "O leilão com id " + idLeilao.Id + " não existe"
				sendMessageToClient(connection, message)
			}
			var message string

			if foundItem.LanceAtual.Email != "" {
				message = vendedor.Nome + " o leilão com o item " + foundItem.Nome + " foi encerrado com sucesso e o vencedor foi " + foundItem.LanceAtual.Email + " com o valor de " + strconv.Itoa(foundItem.LanceAtual.Valor)
			} else {
				message = vendedor.Nome + " o leilão com o item " + foundItem.Nome + " foi encerrado com sucesso mas não teve lances"
			}

			sendMessageToClient(connection, message)
		case "LISTAR_LEILOES":
			var leiloesAEnviar []ItemLeilaoVendedor
			for i, ild := range itensLeilaoDB {
				if ild.Status == LEILAO_ATIVO && ild.IdVendedor == vendedor.Id {
					leiloesAEnviar = append(leiloesAEnviar, ItemLeilaoVendedor{
						Id:   itensLeilaoDB[i].Id,
						Nome: itensLeilaoDB[i].Nome,
					})
				}
			}

			message, _ := json.Marshal(&MessageRespostaListarLeiloesVendedor{
				Leiloes: leiloesAEnviar,
			})
			sendMessageToClient(connection, string(message))
		case "SAIR":
			connection.Close()
			return
		default:
			fmt.Println("Operação não reconhecida")
			message := "Operação não reconhecida"
			sendMessageToClient(connection, message)
		}
	}
}

func handleSocketMessage(connection net.Conn) string {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	handleConnectionError(connection, err, "Perdemos a conexão com o cliente:")
	fmt.Println("Mensagem recebida no socket: ", string(buffer[:mLen]), "do client", connection.RemoteAddr())
	return string(buffer[:mLen])
}

func generateRandomId() string {
	return strconv.Itoa((rand.Intn(1000000)))
}

func sendMessageToClient(connection net.Conn, message string) {
	_, err := connection.Write([]byte(message))
	handleConnectionError(connection, err, "Perdemos a conexão com o cliente:")
}
