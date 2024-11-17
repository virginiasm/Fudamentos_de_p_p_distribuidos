// NOME: RAFAELA V. ALBUQUERQUE E VIRGÍNIA S. MÜLLER
package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [4]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
}

var (
	chans = []chan mensagem{ // vetor de canias para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)

// ElectionController controla o processo de eleição
func ElectionController(in chan int) {
	defer wg.Done()

	var temp mensagem

	// Comandos para o anel iniciam aqui

	// Mudar o processo 3 (canal de entrada 2) para falho (mensagem tipo 2)
	fmt.Println("Controle: mudar o processo 3 para falho")
	temp.tipo = 2
	chans[2] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// Iniciar uma eleição pelo processo 0 (entrada 3)
	fmt.Println("Controle: inicia a eleição")
	temp.tipo = 1
	chans[3] <- temp
	fmt.Printf("Controle: confirmação %d\n", <-in) // Receber e imprimir confirmação

	// Mudar o processo 2 (canal de entrada 1) para falho (mensagem tipo 2)
	fmt.Println("\n\nControle: mudar o processo 2 para falho")
	temp.tipo = 2
	chans[1] <- temp
	fmt.Printf("Controle: confirmação %d\n", <-in) // Receber e imprimir confirmação

	// Iniciar uma eleição pelo processo 0 (entrada 3)
	fmt.Println("Controle: inicia a eleição")
	temp.tipo = 1
	chans[3] <- temp
	fmt.Printf("Controle: confirmação %d\n", <-in) // Receber e imprimir confirmação

	// Mudar o processo 3 (canal de entrada 2) para funcionando (mensagem tipo 3)
	fmt.Println("\n\nControle: mudar o processo 2 para não falho")
	temp.tipo = 3
	chans[2] <- temp
	fmt.Printf("Controle: confirmação %d\n", <-in) // Receber e imprimir confirmação

	// Iniciar uma eleição pelo processo 0 (entrada 3)
	fmt.Println("Controle: inicia a eleição")
	temp.tipo = 1
	chans[3] <- temp
	fmt.Printf("Controle: confirmação %d\n", <-in) // Receber e imprimir confirmação

	// Mudar o processo 3 (canal de entrada 2) para funcionando (mensagem tipo 3)
	fmt.Println("\n\nControle: mudar o processo 3 para não falho")
	temp.tipo = 3
	chans[1] <- temp
	fmt.Printf("Controle: confirmação %d\n", <-in) // Receber e imprimir confirmação

	// Iniciar uma eleição pelo processo 0 (entrada 3)
	fmt.Println("Controle: inicia a eleição")
	temp.tipo = 1
	chans[3] <- temp
	fmt.Printf("Controle: confirmação %d\n", <-in) // Receber e imprimir confirmação

	temp.tipo = 6
	for i := 0; i < len(chans); i++ {
		chans[i] <- temp
	}

	fmt.Println("\n   Processo controlador concluído\n")
}

// ElectionProcess representa um processo no anel de eleição
func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variaveis locais que indicam se este processo é o lider e se esta ativo

	var actualLeader int
	var bFailed bool = false // todos inciam sem falha

	actualLeader = leader // indicação do lider veio por parâmatro

	for temp := range in { // Ler mensagem
		fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])

		if temp.tipo == 6 {
			break
		}

		if bFailed && (temp.tipo != 2 && temp.tipo != 3) {
			fmt.Printf("%2d: estou falhado\n", TaskId)
			out <- temp
			continue
		}

		switch temp.tipo {
		case 1: // Inicia uma eleição
			{
				fmt.Printf("%2d: Estamos iniciando uma eleição\n", TaskId)
				var data mensagem
				data.tipo = 4
				data.corpo[0] = TaskId
				fmt.Printf("%2d: Me candidatei a líder\n", TaskId)

				for i := 1; i < len(data.corpo); i++ {
					data.corpo[i] = -1
				}
				out <- data
				data = <-in

				fmt.Printf("%2d: Recebi mensagem %d, [ %d, %d, %d, %d ]\n", TaskId, data.tipo, data.corpo[0], data.corpo[1], data.corpo[2], data.corpo[3])
				var newLeader int = -1

				for i := 0; i < len(data.corpo); i++ {
					if newLeader < data.corpo[i] {
						newLeader = data.corpo[i]
					}
					data.corpo[i] = 0
				}

				fmt.Printf("%2d: Iniciando processo de contabilização de votos\n", TaskId)

				data.tipo = 5
				data.corpo[0] = newLeader
				actualLeader = newLeader

				fmt.Printf("%2d: O %d foi eleito\n", TaskId, actualLeader)

				out <- data
				data = <-in

				fmt.Printf("%2d: Recebi mensagem %d, [ %d, %d, %d, %d ]\n", TaskId, data.tipo, data.corpo[0], data.corpo[1], data.corpo[2], data.corpo[3])

				fmt.Printf("%2d: novo líder %d\n", TaskId, data.corpo[0])

				controle <- 5
			}
		case 2: // Comando Falhar
			{
				bFailed = true
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
				controle <- -5
			}
		case 3: // Comando Retornar
			{
				bFailed = false
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
				controle <- -5
			}
		case 4: // Comando Escolhe um Líder
			{
				for i := 0; i < len(temp.corpo); i++ {
					if temp.corpo[i] == -1 {
						temp.corpo[i] = TaskId
						break
					}
				}
				fmt.Printf("%2d: Me candidatei a líder\n", TaskId)
				out <- temp
			}
		case 5: // Comando Seta Líder
			{
				actualLeader = temp.corpo[0]
				fmt.Printf("%2d: O %d foi eleito\n", TaskId, actualLeader)
				out <- temp
			}
		default:
			{
				fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			}
		}
	}
	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {
	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[0], 0) // este é o líder
	go ElectionStage(1, chans[0], chans[1], 0) // não é líder, é o processo 1
	go ElectionStage(2, chans[1], chans[2], 0) // não é líder, é o processo 2
	go ElectionStage(3, chans[2], chans[3], 0) // não é líder, é o processo 3

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionController(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Esperar as goroutines terminarem
}
