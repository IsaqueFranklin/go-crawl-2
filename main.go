package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
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
	// Esse concurrentWorkers não é realmente necessário já que o Colly já implementa seu próprio e sofisticado sistema de controle de concorrência e paralelismo internamente.
	//concurrentWorkers = make(chan struct{}, 10) // Limite de 10 workers simultâneos

	// Mutex é usado para proteger, bloquear alguma alteracao num channel quando um goroutine está mexendo naquele channel para que não haja race conditions.
	//O jeito que o Mutex faz isso é simplesmente sincronizando o recebedor e o rementente de um dado do channel.
	//Nesse caso o Mutex é para proteger o mapa de URLs visitadas e a slice de URLs encontradas.
	mu sync.Mutex

	// Conjunto para armazenar URLs já visitadas para evitar processamento duplicado
	visitedURLs = make(map[string]bool)

	// Slice para armazenar todas as URLs de marketing encontradas
	marketingURLs []ScrapedURL

	// Regex (expressões regulares) para filtar URLs de marketing
	marketingKeywords = regexp.MustCompile(`(?i)(marketing|blog|content|digital|seo|sem|social|inbound|outbound|growth|strategy|conversion|branding)`)

	// Domínios iniciais para o crawler. Lista expandida!
	seedURLs = []string{
		"https://neilpatel.com/blog/",
		"https://moz.com/blog",
		"https://www.hubspot.com/marketing",
		"https://blog.rockcontent.com/br/",
		"https://www.searchenginejournal.com/",
		"https://contentmarketinginstitute.com/",
		"https://adespresso.com/blog/",
		"https://blog.hootsuite.com/",
		"https://blog.rdstation.com/",
		"https://resultadosdigitais.com.br/blog/",
		"https://www.semrush.com/blog/",
		"https://marketingdeconteudo.com/",
		"https://klickpages.com.br/blog/",
		"https://www.vtex.com/pt-br/blog/",
		"https://ecommercenapratica.com/blog/",
		"https://shopify.com.br/blog/",
		"https://marketing.substack.com/",
		"https://growthhackers.com/blog/",
		"https://www.martechalliance.com/blog",
		"https://blog.agenciaeplus.com.br/",
		"https://www.mktdigital.com.br/blog/",
		"https://www.ecommercebrasil.com.br/artigos/",
		"https://ecommercefluente.com.br/",
		"https://mundodomarketing.com.br/",
		"https://sebrae.com.br/sites/PortalSebrae/cursosonline/como-fazer-marketing-digital-para-sua-empresa,2a9fe47f1c070410VgnVCM1000004c00210aRCRD",
		"https://www.hostgator.com.br/blog/marketing-digital/",
		"https://digitalhouse.com/br/blog/marketing-digital/",
		"https://www.alura.com.br/artigos/marketing-digital",
		"https://www.ecommerce.org.br/artigos",
		"https://conradoadolpho.com/blog/",
	}

	// Domínios para os quais o crawler está permitido a seguir links
	// Isso evita que o crawler vá para domínios completamente não relacionados
	allowedDomains = []string{
		"neilpatel.com",
		"moz.com",
		"hubspot.com",
		"blog.rockcontent.com",
		"searchenginejournal.com",
		"contentmarketinginstitute.com",
		"adespresso.com",
		"blog.hootsuite.com",
		"blog.rdstation.com",
		"resultadosdigitais.com.br",
		"semrush.com",
		"marketingdeconteudo.com",
		"klickpages.com.br",
		"vtex.com",
		"ecommercenapratica.com",
		"shopify.com.br",
		"marketing.substack.com",
		"growthhackers.com",
		"martechalliance.com",
		"blog.agenciaeplus.com.br",
		"mktdigital.com.br",
		"ecommercebrasil.com.br",
		"ecommercefluente.com.br",
		"mundodomarketing.com.br",
		"sebrae.com.br",
		"hostgator.com.br",
		"digitalhouse.com",
		"alura.com.br",
		"ecommerce.org.br",
		"conradoadolpho.com",
	}
)

func main() {
	fmt.Println("Starting the Web Crawler for Marketing websites...")

	// Inicializa o coletor Colly
	c := colly.NewCollector(
		// Permite a recursão (seguir links)
		colly.AllowURLRevisit(),

		colly.Async(true), // Esta opção instrui o Colly a usar goroutines para processar as requisições de forma assíncrona, gerenciando a fila de trabalho e a execução paralela de forma eficiente.

		// User-Agent para simular um navegador real
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"),
	)

	// Configuracoes de limite de concorrência para o coletor
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",              // Aplica a todos os domínios.
		Parallelism: 10,               // 10 requisicoes paralelas por domínio.
		Delay:       10 * time.Second, // Define o atrase entre as requisicões para evitar ser bloqueado.
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
		parsedURL, err := url.Parse(absoluteURL) // Tenta analisar (parsear) a URL. Se ela for malformada ou inválida (err != nil), a função simplesmente ignora esse link (return).
		if err != nil {
			return
		}
		fmt.Println("Eis o parsedURL: ", parsedURL)
		isValidDomain := false
		for _, domain := range allowedDomains {
			if strings.Contains(parsedURL.Hostname(), domain) {
				isValidDomain = true
				break
			}
		}

		if !isValidDomain {
			return // Ignora links para domínios não permitidos. Isso evita que seu crawler "fuja" do escopo e comece a vasculhar a internet inteira.
		}

		// Verifica se a URL contém palavras-chave de marketing e não é um link para arquivos
		if marketingKeywords.MatchString(absoluteURL) && !strings.Contains(absoluteURL, ".pdf") && !strings.Contains(absoluteURL, ".zip") && !strings.Contains(absoluteURL, ".doc") {
			mu.Lock()
			marketingURLs = append(marketingURLs, ScrapedURL{
				URL:         absoluteURL,
				Source:      e.Request.URL.Hostname(),
				FoundOnPage: e.Request.URL.String(),
				Timestamp:   time.Now().Format(time.RFC3339),
			})
			mu.Unlock()
		}

		// Visita o link se ele for relevante e não for um arquivo estático
		// Colly lida com a lógica de recursão e visita de links
		if !strings.Contains(absoluteURL, ".css") && !strings.Contains(absoluteURL, ".js") && !strings.Contains(absoluteURL, ".png") && !strings.Contains(absoluteURL, ".jpg") && !strings.Contains(absoluteURL, ".gif") {
			e.Request.Visit(link) //Caso todos os requisitos batam o colly visita o link e encontra novas possíveis urls para recomcar o ciclo, daí a recursão que o Colly resolve sozinho.
		}
	})

	// OnError é chamado se ocorrer um erro durante a requisicão
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Printf("Error when visiting: %v\n", err)
	})

	// Loop através das URLs iniciais e inicia o crawling.
	for _, u := range seedURLs {
		c.Visit(u)
	}

	// Espera até que todas as goroutines do Colly terminem antes que o programa principal continue.
	c.Wait()

	fmt.Printf("\nCrawiling finished. It found %d URLs with marketing content.\n", len(marketingURLs))

	// Salva as URLs coletadas em um arquivo JSON
	if err := saveURLs(marketingURLs, "marketing_urls.json"); err != nil {
		fmt.Printf("Error saving URLs: %v\n", err)
	} else {
		fmt.Println("URLs saved in marketing_urls.json file.")
	}
}

// saveURLs salva a slice de ScrapedURL em um arquivo JSON.
func saveURLs(urls []ScrapedURL, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("there was an error creating file %s: %w", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Formatacão bonita do JSON
	if err := encoder.Encode(urls); err != nil {
		return fmt.Errorf("there was an error enconding the URLs to JSON: %w", err)
	}
	return nil
}
