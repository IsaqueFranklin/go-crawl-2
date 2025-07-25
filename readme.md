# 📁 Estrutura do Projeto de Crawler em Go

## 🔧 Funções / Arquivos

- **`main`**:  
  Ponto de entrada do programa, onde tudo é inicializado.

- **`crawl`**:  
  Função principal responsável por:
  - Fazer a requisição HTTP
  - Fazer o parse do HTML
  - Extrair os links da página

- **`worker`**:  
  Função executada por cada goroutine para processar as URLs da fila (concorrência).

- **`addToQueue`**:  
  Adiciona uma nova URL à fila de URLs a serem visitadas, evitando duplicatas.

- **`saveURLs`**:  
  Salva as URLs coletadas em um arquivo de saída.

---

## 📚 Bibliotecas Utilizadas

- **`colly`**:  
  Usada para simplificar a lógica de crawling. Essa biblioteca já cuida de muitas complexidades como:
  - Gerenciamento de links visitados
  - Execução concorrente
  - Requisições HTTP
  - Parse de páginas

- **`goquery`**:  
  Embora o `colly` tenha seu próprio parser, o `goquery` pode ser útil para parsing HTML mais complexo.  
  Ele permite manipular o DOM de forma similar ao jQuery.

---

> ✅ Com essa estrutura, o projeto se mantém organizado, modular e fácil de expandir.
