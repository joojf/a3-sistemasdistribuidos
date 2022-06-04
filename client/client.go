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

type Client struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Artigo struct {
	Descricao   string `json:"descricao"`
	LanceMinimo string `json:"lanceMinimo"`
}

type Request struct {
	Params string `json:"params"`
	Valor  []byte `json:"valor"`
}

type Leilao struct {
	Id     string
	Artigo Artigo
}

type ResponseListarLeiloes struct {
	Leiloes []Leilao `json:"leiloes"`
}

type RequestFazerLance struct {
	IdLeilao string `json:"idLeilao"`
	Valor    int `json:"valor"`
}

type RequestEncerrarLeilao struct {
	Id string `json:"id"`
}

func main() {
	nome, email, role := getUserInfo()

	if role == "Vendedor" {
		criarVendedor(nome, email)
	} else {
		criarComprador(nome, email)
	}
}

func getUserInfo() (string, string, string) {
	nome := promptui.Prompt{
		Label: "Nome",
	}
	email := promptui.Prompt{
		Label: "Email",
	}
	role := promptui.Select{
		Label: "Role",
		Items: []string{"Vendedor", "Comprador"},
	}

	returnedNome, err := nome.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	returnedEmail, err := email.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, resultReturnedRole, err := role.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return returnedNome, returnedEmail, resultReturnedRole
}

func criarVendedor(nome string, email string) {
	connection, err := net.Dial(CONNECTION_TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer connection.Close()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Erro ao criar vendedor: ", r)
		}
	}()

	for {
		prompt := promptui.Select{
			Label: "Escolha uma opção",
			Items: []string{"Criar Leilão", "Encerrar Leilão", "Sair"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch result {
		case "Criar Leilão":
			descricao, lanceMinimo := getLeilaoInfo(connection)
			artigo, _ := json.Marshal(Artigo{
				Descricao:   descricao,
				LanceMinimo: strconv.Itoa(lanceMinimo),
			})

			request, _ := json.Marshal(Request{
				Params: "criarLeilao",
				Valor:  artigo,
			})

			connection.Write(request)

			response := make([]byte, 1024)
			connection.Read(response)

			var responseLeilao Leilao
			json.Unmarshal(response, &responseLeilao)

			fmt.Println("Leilão criado com sucesso!")
			fmt.Println("ID: ", responseLeilao.Id)
			return
		case "Encerrar Leilão":
			responseLeiloes := listarLeiloes(connection)
			leiloes := responseLeiloes.Leiloes

			if len(leiloes) == 0 {
				fmt.Println("Não existem leilões para encerrar!")
				return
			}

			leilao := promptui.Select{
				Label: "Escolha o leilão",
				Items: leiloes,
			}

			i, _, err := leilao.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			request, _ := json.Marshal(RequestEncerrarLeilao{
				Id: leiloes[i].Id,
			})

			encerrarLeilao, _ := json.Marshal(Request{
				Params: "encerrarLeilao",
				Valor:  request,
			})

			connection.Write(encerrarLeilao)

			response := make([]byte, 1024)
			connection.Read(response)

			fmt.Println("Leilão encerrado com sucesso!")
		case "Sair":
			os.Exit(0)
		}
	}
}

func criarComprador(nome string, email string) {
	connection, err := net.Dial(CONNECTION_TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer connection.Close()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Erro ao criar comprador: ", r)
		}
	}()

	for {
		prompt := promptui.Select{
			Label: "Escolha uma opção",
			Items: []string{"Listar Leilões", "Fazer Lance", "Sair"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch result {
		case "Listar Leilões":
			responseLeiloes := listarLeiloes(connection)
			leiloes := responseLeiloes.Leiloes

			if len(leiloes) == 0 {
				fmt.Println("Não existem leilões para comprar!")
				return
			}

			for i, leilao := range leiloes {
				fmt.Printf("%d - %s\n", i, leilao.Artigo.Descricao)
			}

			return
		case "Fazer Lance":
			idLeilao, valor := getLanceInfo(connection)

			lance, _ := json.Marshal(RequestFazerLance{
				IdLeilao: idLeilao,
				Valor:    valor,
			})

			request, _ := json.Marshal(Request{
				Params: "fazerLance",
				Valor:  lance,
			})

			connection.Write(request)

			response := make([]byte, 1024)
			connection.Read(response)

			fmt.Println("Lance feito com sucesso!")
		case "Sair":
			os.Exit(0)
		}
	}
}

func listarLeiloes(connection net.Conn) ResponseListarLeiloes {
	request, _ := json.Marshal(Request{
		Params: "listarLeiloes",
	})

	connection.Write(request)

	response := make([]byte, 1024)
	connection.Read(response)

	var responseLeiloes ResponseListarLeiloes
	json.Unmarshal(response, &responseLeiloes)

	return responseLeiloes
}

func getLeilaoInfo(connection net.Conn) (string, int) {
	descricao := promptui.Prompt{
		Label: "Digite a descrição do leilão",
	}

	descricaoLeilao, _ := descricao.Run()

	lanceMinimo := promptui.Prompt{
		Label: "Digite o lance mínimo",
	}

	lanceMinimoLeilao, _ := lanceMinimo.Run()

	valor, _ := strconv.Atoi(lanceMinimoLeilao)

	return descricaoLeilao, valor
}

func getLanceInfo(connection net.Conn) (string, int) {
	idLeilao := promptui.Prompt{
		Label: "Digite o id do leilão",
	}

	idLeilaoLeilao, _ := idLeilao.Run()

	lanceMinimo := promptui.Prompt{
		Label: "Digite o lance mínimo",
	}

	lanceMinimoLeilao, _ := lanceMinimo.Run()

	valor, _ := strconv.Atoi(lanceMinimoLeilao)

	return idLeilaoLeilao, valor
}