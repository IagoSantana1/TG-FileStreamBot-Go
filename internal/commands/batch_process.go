package commands

import "fmt"

func processBatch(chatID int64, items []BatchItem) {

	fmt.Println("-----------------------------------")
	fmt.Println("Novo lote")
	fmt.Println("Chat:", chatID)
	fmt.Println("Arquivos:", len(items))
	fmt.Println("-----------------------------------")

}
