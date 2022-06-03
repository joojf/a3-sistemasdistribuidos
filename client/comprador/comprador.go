package main

import (
	"encoding/json"
	"fmt"
	"log"
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

type Comprador struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Mensagem struct {
	Operacao string `json:"operacao"`
	Valor    []byte `json:"valor"`
}

type ResponseListaLeiloes struct {
	Leiloes []Leilao `json:"leiloes"`
}

type Leilao struct {
	Id              int
	Nome            string
	LanceMinimo     int
	DescricaoArtigo string
	Vencedor        string
}

type RequestLance struct {
	IdLeilao int
	Email    string
	Valor    float64
}

func main() {
	connection, err := net.Dial(CONNECTION_TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}

	nome, email := obterDadosUsuario()

	comprador, _ := json.Marshal(Comprador{Nome: nome, Email: email, Role: "comprador"})

	_, err = connection.Write(comprador)
	if err != nil {
		log.Fatal(err)
	}

	defer connection.Close()
	defer func() {
		fmt.Println("Encerrando conexão...")
	}()

	for {
		prompt := promptui.Select{
			Label: "Escolha uma opção",
			Items: []string{"Listar Leilões", "Fazer um lance", "Sair"},
		}

		_, result, err := prompt.Run()

		if err != nil {
			log.Fatal(err)
		}

		handleResult(result, connection)
	}
}

func handleResult(result string, connection net.Conn) {
	switch result {
	case "Listar Leilões":
		listarLeiloes(connection)
	case "Fazer um lance":
		fazerLance(connection)
	case "Sair":
		os.Exit(0)
	}
}

func listarLeiloes(connection net.Conn) {
	fmt.Println("Listando leilões...")

	mensagem, err := json.Marshal(Mensagem{Operacao: "listarLeiloes"})

	_, err = connection.Write(mensagem)
	if err != nil {
		log.Fatal(err)
	}

	var response ResponseListaLeiloes
	err = json.NewDecoder(connection).Decode(&response)
	if err != nil {
		log.Fatal(err)
	}

	for _, leilao := range response.Leiloes {
		fmt.Printf("%d - %s - %d - %s\n", leilao.Id, leilao.Nome, leilao.LanceMinimo, leilao.DescricaoArtigo)
	}
}

func fazerLance(connection net.Conn) {
	fmt.Println("Fazendo um lance...")

	email := obterEmailUsuario()

	leiloes, err := json.Marshal(Mensagem{Operacao: "listarLeiloes"})

	_, err = connection.Write(leiloes)
	if err != nil {
		log.Fatal(err)
	}

	var response ResponseListaLeiloes
	err = json.NewDecoder(connection).Decode(&response)
	if err != nil {
		log.Fatal(err)
	}

	leiloesDisponiveis := []string{}
	for _, leilao := range response.Leiloes {
		leiloesDisponiveis = append(leiloesDisponiveis, strconv.Itoa(leilao.Id))
	}

	prompt := promptui.Select{
		Label: "Escolha o leilão",
		Items: leiloesDisponiveis,
	}

	_, result, err := prompt.Run()

	if err != nil {
		log.Fatal(err)
	}

	idLeilao, _ := strconv.Atoi(result)

	promptValor := promptui.Prompt{
		Label: "Digite o valor do lance: ",
	}

	valor, err := promptValor.Run()

	if err != nil {
		log.Fatal(err)
	}

	valorFloat, err := strconv.ParseFloat(valor, 64)

	lance, err := json.Marshal(RequestLance{
		IdLeilao: idLeilao,
		Email:    email,
		Valor:    valorFloat,
	})

	_, err = connection.Write(lance)
	if err != nil {
		log.Fatal(err)
	}
}

func obterDadosUsuario() (string, string) {
	promptNome := promptui.Prompt{
		Label: "Digite seu nome: ",
	}

	nome, err := promptNome.Run()

	if err != nil {
		log.Fatal(err)
	}

	promptEmail := promptui.Prompt{
		Label: "Digite seu email: ",
	}

	email, err := promptEmail.Run()

	if err != nil {
		log.Fatal(err)
	}

	return nome, email
}

func obterEmailUsuario() string {
	promptEmail := promptui.Prompt{
		Label: "Digite seu email: ",
	}

	email, err := promptEmail.Run()

	if err != nil {
		log.Fatal(err)
	}

	return email
}
