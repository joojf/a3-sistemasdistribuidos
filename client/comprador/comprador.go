package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

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
	Id           string `json:"id"`
	Nome         string `json:"nome"`
	Descricao    string `json:"descricao"`
	ValorInicial int    `json:"valorInicial"`
	LanceAtual   Lance  `json:"lanceAtual"`
}

type Lance struct {
	Email string `json:"Email"`
	Valor int    `json:"valor"`
}

type MessageListaDeLeiloes struct {
	Leiloes []ItemLeilaoCliente `json:"leiloes"`
}
type MessageEncerrarLeilao struct {
	Id string `json:"id"`
}

type MessageDarLance struct {
	Id    string `json:"id"`
	Valor int    `json:"valor"`
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
		Role:  "comprador",
	})

	_, err = connection.Write(vendedor)
	if err != nil {
		fmt.Println("Erro ao enviar mensagem para o servidor")
	}
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Erro ao receber mensagem do servidor")
	}
	fmt.Println("Resposta do servidor:", string(buffer[:mLen]))

	defer connection.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Ocorreu um erro na conexão do cliente ou cliente se desconectou: ", err)
		}
	}()
	for {

		prompt := promptui.Select{
			Label: "Selecione a operação",
			Items: []string{"Listar Artigos", "Fazer um lance", "Sair"},
		}
		_, result, err := prompt.Run()

		handleError(err, "Erro ao selecionar opção: %v\n")
		if err != nil {
			continue
		}
		handleUserResponse(result, connection)
	}
}

func handleUserResponse(response string, connection net.Conn) {
	switch response {
	case "Listar Artigos":
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

		for _, leilao := range listaLeiloes {
			var maiorLance string
			maiorLance = "Sem lances"
			if (leilao.LanceAtual != Lance{}) {
				maiorLance = strconv.Itoa(leilao.LanceAtual.Valor)
			}
			fmt.Printf("Nome: %s\nDescrição: %s\nValor inicial: %d\nMaior lance: %s\n", leilao.Nome, leilao.Descricao, leilao.ValorInicial, maiorLance)
			fmt.Printf("-----------------------------------------------------\n")
		}
		return
	case "Fazer um lance":
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
			Label: "Selecione o leilão a dar o lance",
			Items: listaLeiloes,
		}
		i, _, err := prompt.Run()

		handleError(err, "Erro ao selecionar lance: %v\n")
		if err != nil {
			return
		}

		promptValorLance := promptui.Prompt{
			Label: "Lance",
		}

		valor, err1 := promptValorLance.Run()
		handleError(err1, "Erro ao colocar valor do lance: %v\n")
		if err1 != nil {
			return
		}

		valorConvertido, err2 := strconv.Atoi(valor)
		handleError(err1, "Erro ao converter valor do lance: %v\n")

		if err2 != nil {
			return
		}

		lanceLeilao, _ := json.Marshal(&MessageDarLance{
			Valor: int(valorConvertido),
			Id:    listaLeiloes[i].Id,
		})

		messageDarLance, _ := json.Marshal(&Message{
			Operacao: "DAR_LANCE",
			Message:  lanceLeilao,
		})
		sendMessageToServer(connection, messageDarLance, "Erro ao dar lance: %v\n")

		receivedDarLanceMessage := receiveMessageFromServer(connection)
		fmt.Print(receivedDarLanceMessage + "\n")
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
	handleError(err1, "Erro ao colocar nome: %v\n")

	email, err2 := promptEmail.Run()
	handleError(err2, "Erro ao colocar email: %v\n")

	return nome, email
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
