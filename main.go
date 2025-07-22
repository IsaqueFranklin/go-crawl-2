package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// Essa é a struct que vai representar uma URL coletada
type ScrapedURL struct {
	URL         string `json:"url"`
	Source      string `json:"source"`
	FoundOnPage string `json:"found_on_page"`
	Timestamp   string `json:"timestamp"`
}

// IMPORTANTE: Em Go, o bloco var ( ... ) é apenas uma forma de declarar várias variáveis de uma vez só, em vez de escrever var toda hora.
var (
	// Esse é um semáforo para limitar o número de workers concorrentes
	//isso evita sobrecarregar o site alvo ou minha máquina
	concurrentWorkers = make(chan struct{}, 10) // Limite de 10 workers simultâneos

	// Mutex é usado para proteger, bloquear alguma alteracao num channel quando um goroutine está mexendo naquele channel para que não haja race conditions.
	//O jeito que o Mutex faz isso é simplesmente sincronizando o recebedor e o rementente de um dado do channel.
	//Nesse caso o Mutex é para proteger o mapa de URLs visitadas e a slice de URLs encontradas.
	mu sync.Mutex

	// Conjunto para armazenar URLs já visitadas para evitar processamento duplicado
	visitedURLs = make(map[string]bool)

	// Regex (expressões regulares) para filtar URLs de marketing
	marketingKeywords = regexp.MustCompile(`(?i)(marketing|blog|content|digital|seo|sem|social|inbound|outbound|growth|strategy|conversion|branding)`)

	// Domínios iniciais para o crawler
	seedURLs = []string{
		"neilpatel.com",
		"moz.com",
		"hubspot.com",
		"rockcontent.com",
		"searchenginejournal.com",
	}

	// Domínios para os quais o crawler está permitido a seguir links
	// Isso evita que o crawler vá para domínios completamente não relacionados
	allowedDomains = []string{
		"neilpatel.com",
		"moz.com",
		"hubspot.com",
		"rockcontent.com",
		"searchenginejournal.com",
	}
)

func main() {
	fmt.Println("Starting the Web Crawler for Marketing websites...")

	// Inicializa o coletor Colly
	c := colly.NewCollector(
		// Permite a recursão (seguir links)
		colly.AllowURLRevisit(),
		// Limita o número de threads paralelas. Colly gerencia isso internamente com goroutines.
		colly.Async(true),
		// User-Agent para simular um navegador real
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebkit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"),
	)

	// Configuracoes de limite de concorrência para o coletor
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",             // Aplica a todos os domínios
		Parallelism: 5,               // 5 requisicoes paralelas por domínio
		Delay:       1 * time.Second, // Define o atrase entre as requisicões para evitar ser bloqueado.
	})

	// Define os domínios permitidos
	c.AllowedDomains = allowedDomains

	// Callbacks Colly:

	// OnRequest é chamado antes de cada requisicao HTTP
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting: %s\n", r.URL.String())
	})

	// OnHTML é chamado quando um elemento HTML específico é encontrado
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		// Verifica se a URL já foi visitada
		mu.Lock()
		if visitedURLs[absoluteURL] {
			mu.Unlock()
			return
		}
		visitedURLs[absoluteURL] = true
		mu.Unlock()

		// Parse da URL para verificar se é válida e dentro dos domínios permitidos
		parsedURL, err := url.Parse(absoluteURL)
		if err != nil {
			return
		}

		isValidDomain := false
		for _, domain := range allowedDomains {
			if strings.Contains(parsedURL.Hostname(), domain) {
				isValidDomain = true
				break
			}
		}

		if !isValidDomain {
			return // Ignora links para domínios não permitidos.
		}
	})
}
