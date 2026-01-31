# Inicializar server

- Execute `go run ./server` para iniciar o servidor na porta 8080.

# Inicializar client

- Execute `go run ./client` para iniciar o cliente que se conectará ao servidor na porta 8080.

# Testar comunicação

- O server deverá logar o status da requisição, o response, o erro, e o momento que a request finalizou.
- O client deverá escrever no arquivo `cotacao.txt` o valor retornado pelo servidor no formato esperado.
- O arquivo `cotacoes.db` deve poder ser aberto em uma ferramenta como https://inloop.github.io/sqlite-viewer/ que permita visualizar o conteúdo do banco de dados SQLite.
