=> SOBRE AS FUNCOES/FILES

- main: É o ponto de entrada do programa, onde inicializamos tudo.
- crawl: A funcao principal que fará a requisicao, parse do HTML e extracao de links.
- worker: Uma funcao que será executada por cada goroutine para processar as URLs da fila.
- addToQUeue: Adiciona uma URL à fila de URLs a serem visitadas.
- saveURLs: Salva as URLs coletadas em um arquivo.

=> SOBRE AS LIBRARIES USADAS

- colly: Vamos usar colly para simplificar a lógica de crawling, pois ele já cuida de muitas complexidades, como gerenciamente de links visitados e concorrência.
- goquery: Embora colly use sua própria lógica, goquery é bom para parsing mais complexo, mas colly é mais alto nível.