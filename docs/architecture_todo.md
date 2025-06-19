# 🛠️ Architecture Refactor TODO

Este documento registra ideias futuras que serão implementadas gradualmente no projeto, com o objetivo de elevar a qualidade de código, organização e escalabilidade.

---

## 1. Inversão de Dependência via Interfaces

* Criar interfaces como `PostRepository`, `PostStorage`, etc.
* Ter implementações concretas como `PostgresPostRepository`, `InMemoryPostRepository`.
* Services devem depender de interfaces, não de implementações concretas.

## 2. Função Centralizada de Parsing e Validação

* Criar uma função genérica:

  ```go
  func ParseAndValidate[T any](r *http.Request, dst *T) error
  ```
* Usa `json.Decoder` + `go-playground/validator`
* Local: `pkg/httpx/json.go` ou `internal/utils/json.go`

## 3. Conversão do DTO para Model (Normalize Early)

* Adicionar método ao DTO:

  ```go
  func (dto CreatePostRequest) ToModel() (model.Post, error)
  ```
* Handler chama isso direto e envia para o service.

## 4. Regra do `publishedAt` no Service

* Handler apenas envia o booleano `Published`
* Service define se deve preencher o `PublishedAt`

## 5. Função para Gerar `PostResponse`

* Criar:

  ```go
  func FromPostModel(post model.Post) PostResponse
  ```
* Centraliza conversão do model para resposta JSON.

## 6. Funções Centralizadas para Respostas HTTP

* `httpx.WriteJSON(w, status, data)`
* `httpx.WriteError(w, status, message)`
* `httpx.WriteValidationErrors(w, errs)`
* Local: `pkg/httpx/response.go`

## 7. Função dedicada para parsing

* Criar:

  ```go
  req, err := httpx.Bind[CreatePostRequest](r)
  ```
* Uso de generics + validação simplifica handlers.

## 8. Uso de `context.Context`

* Começar a propagar `ctx` da requisição para todas as camadas:

  ```go
  func (s *PostService) CreatePost(ctx context.Context, post model.Post) (*model.Post, error)
  ```
* Permite timeout, cancelamento, rastreamento etc.

---

## 📁 Estrutura recomendada para organização futura

```
internal/
  modules/
    posts/
      dto/
        post_dto.go
      delivery/
        post_handler.go
      service/
        post_service.go
      repository/
        interface.go
        postgres.go
pkg/
  httpx/
    json.go
    response.go
docs/
  architecture_todo.md
```

---

Essas melhorias serão aplicadas gradualmente conforme o MVP evolui. O objetivo é manter o código limpo, testável e escalável com qualidade de produção.
