package main

import (
	"fmt"
	"sync"
)

func race_condition() {
	var count int
	var mu sync.Mutex // sem o mu e mu.Lock/Unlock o resultado é menos de 1000

	// O sync.WaitGroup é usado para esperar que múltiplas goroutines terminem antes de continuar o programa.
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		// wg.Add(n) Diz ao WaitGroup: "Vou lançar n goroutines que devem ser aguardadas", nesse caso n = 1.
		wg.Add(1)
		go func() {
			// wg.Done() Diz ao WaitGroup: "Uma das goroutines terminou".
			defer wg.Done() // Nesse o defer espera a goroutine retornar para aí sim dar um Done no wg e avisar que acabou.
			mu.Lock()       // Bloqueia acesso exclusivo à seção crítica (no caso o count++).
			count++
			mu.Unlock() // Libera o mutex para outra goroutine.
		}()
	}

	// wg.Wait() Espera até que todas as goroutines façam Done() antes de seguir o programa.
	wg.Wait()
	fmt.Println("Resultado final:", count)
}
