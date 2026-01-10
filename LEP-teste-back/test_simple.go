package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func TestSimple() {
	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Println("TESTE SIMPLES DE CONECTIVIDADE")
	fmt.Println(separator + "\n")

	// Teste 1: Verificar se backend está rodando
	fmt.Print("→ Testando conexão em http://localhost:8080... ")

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("http://localhost:8080/ping")
	if err != nil {
		fmt.Printf("\n❌ ERRO: Backend não está respondendo\n")
		fmt.Printf("   Detalhes: %v\n\n", err)
		fmt.Println("SOLUÇÃO:")
		fmt.Println("  1. Abra outro terminal")
		fmt.Println("  2. Execute: cd LEP-Back && go run main.go")
		fmt.Println("  3. Aguarde mensagem: 'Server running on :8080'")
		fmt.Println("  4. Volte para este terminal e execute novamente\n")
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ OK (Status: %d)\n\n", resp.StatusCode)
	fmt.Println("Backend está respondendo corretamente!")
	fmt.Println("Tudo pronto para executar os testes completos!")
	fmt.Println("\nPróximo passo: Execute 'go run .' sem argumentos\n")
}
