package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/manifoldco/promptui"
)

const (
	HOST            = "localhost"
	PORT            = "12345"
	CONNECTION_TYPE = "tcp"
)

type Vendedor struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ItemLeilao struct {
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Valor     string `json:"valor"`
}

type Message struct {
	Operacao string `json:"operacao"`
	Message  []byte `json:"message"`
}

type ItemLeilaoCliente struct {
	Id   string
	Nome string
}
type MessageListaDeLeiloes struct {
	Leiloes []ItemLeilaoCliente `json:"leiloes"`
}
type MessageEncerrarLeilao struct {
	Id string `json:"id"`
}

func main() {
	connection, err := net.Dial(CONNECTION_TYPE, HOST+":"+PORT)
	if err != nil {
		panic(err)
	}

	nome, email := promptCredentials()

	vendedor, _ := json.Marshal(&Vendedor{
		Nome:  nome,
		Email: email,
		Role:  "vendedor",
	})

	_, err = connection.Write(vendedor)
	if err != nil {
		fmt.Println("Erro ao enviar credenciais: %v\n", err.Error())
	}
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Erro ao receber credenciais: %v\n", err.Error())
	}
	fmt.Println("Resposta do servidor: " + string(buffer[:mLen]))

	defer connection.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Ocorreu um erro na conexão do cliente ou cliente se desconectou: ", err)
		}
	}()
	for {

		prompt := promptui.Select{
			Label: "Selecione a operação",
			Items: []string{"Iniciar Leilão", "Encerrar Leilão", "Sair"},
		}
		_, result, err := prompt.Run()

		handleError(err, "Erro ao selecionar opção: %v\n")
		handleUserResponse(result, connection)
	}
}

func handleUserResponse(response string, connection net.Conn) {
	switch response {
	case "Iniciar Leilão":
		nome, descricao, valorInicial := promptAuctionDetails()
		item, _ := json.Marshal(&ItemLeilao{
			Nome:      nome,
			Descricao: descricao,
			Valor:     valorInicial,
		})
		message, _ := json.Marshal(&Message{
			Operacao: "CRIAR_LEILAO",
			Message:  item,
		})
		sendMessageToServer(connection, message, "Erro ao criar leilão: %v\n")
		receivedMessage := receiveMessageFromServer(connection)
		fmt.Println("Resposta do servidor: " + receivedMessage)
		return
	case "Encerrar Leilão":
		messageListarLeiloes, _ := json.Marshal(&Message{
			Operacao: "LISTAR_LEILOES",
			Message:  make([]byte, 0),
		})
		sendMessageToServer(connection, messageListarLeiloes, "Erro ao listar leilões: %v\n")
		receivedLeiloesMessage := receiveMessageFromServer(connection)
		var jsonMsg MessageListaDeLeiloes
		json.Unmarshal([]byte(receivedLeiloesMessage), &jsonMsg)
		listaLeiloes := jsonMsg.Leiloes
		if len(listaLeiloes) == 0 {
			fmt.Println("Não há leilões disponíveis")
			return
		}
		prompt := promptui.Select{
			Label: "Selecione o leilão a encerrar",
			Items: listaLeiloes,
		}
		i, _, err := prompt.Run()

		handleError(err, "Erro ao encerrar leilão: %v\n")

		idLeilao, _ := json.Marshal(&MessageEncerrarLeilao{
			Id: listaLeiloes[i].Id,
		})

		messageEncerrarLeilao, _ := json.Marshal(&Message{
			Operacao: "ENCERRAR_LEILAO",
			Message:  idLeilao,
		})
		sendMessageToServer(connection, messageEncerrarLeilao, "Erro ao encerrar leilão: %v\n")

		receivedEncerramentoMessage := receiveMessageFromServer(connection)
		fmt.Print(receivedEncerramentoMessage + "\n")
		return
	case "Sair":
		messageEncerrarLeilao, _ := json.Marshal(&Message{
			Operacao: "SAIR",
			Message:  make([]byte, 0),
		})
		sendMessageToServer(connection, messageEncerrarLeilao, "Erro ao encerrar leilão: %v\n")
		os.Exit(0)
	}
}

func promptCredentials() (nome, email string) {
	promptNome := promptui.Prompt{
		Label: "Nome",
	}
	promptEmail := promptui.Prompt{
		Label: "Email",
	}

	nome, err1 := promptNome.Run()
	handleError(err1, "Erro ao obter nome: %v\n")

	email, err2 := promptEmail.Run()
	handleError(err2, "Erro ao obter email: %v\n")

	return nome, email
}

func promptAuctionDetails() (nome, descricao string, valorInicial string) {
	promptNome := promptui.Prompt{
		Label: "Nome do artigo",
	}
	promptDescricao := promptui.Prompt{
		Label: "Descrição do artigo",
	}
	promptValor := promptui.Prompt{
		Label: "Valor inicial",
	}

	nome, err1 := promptNome.Run()
	handleError(err1, "Erro ao obter nome: %v\n")

	email, err2 := promptDescricao.Run()
	handleError(err2, "Erro ao obter email: %v\n")

	descricao, err3 := promptValor.Run()
	handleError(err3, "Erro ao obter email: %v\n")

	return nome, email, descricao
}

func handleError(err error, message string) {
	if err != nil {
		fmt.Printf(message, err.Error())
		panic(err)
	}
}

func handleConnectionError(connection net.Conn, err error, message string) {
	if err != nil {
		handleError(err, message)
		connection.Close()
		panic(err)
	}
}

func sendMessageToServer(connection net.Conn, message []byte, errorMessage string) {
	_, err := connection.Write(message)
	handleError(err, errorMessage)
}

func receiveMessageFromServer(connection net.Conn) string {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	handleConnectionError(connection, err, "Perdemos a conexão com o servidor")
	return string(buffer[:mLen])
}
