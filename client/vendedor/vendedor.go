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

type Vendedor struct {
	Nome  string
	Email string
	Role  string
}

type Cliente struct {
	Id    int
