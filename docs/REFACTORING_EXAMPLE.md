# 🔧 Refatoração: Eliminando Vazamento de *gorm.DB

## ❌ ANTES (Vazando infraestrutura)

```go
// internal/app/router/app/products.go
func NewProductRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
    dbService := rtl.GetDatabaseService()
    dbGorm, err := dbService.GetDB()  // ❌ VAZOU *gorm.DB

    // ❌ Controller precisa receber *gorm.DB
    productRepo := gdbasez.NewProductRepo(dbGorm)  // ❌ VAZA
    productService := gdbasez.NewProductService(productRepo)
    productController := products.NewProductController(productService)

    // ... routes
}
```

## ✅ DEPOIS (Bridge Limpa)

```go
// internal/app/router/app/products.go
func NewProductRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
    dbService := rtl.GetDatabaseService()
    dbGorm, err := dbService.GetDB()

    // ✅ CRIAR BRIDGE (único ponto onde *gorm.DB aparece)
    bridge := gdbasez.NewBridge(dbGorm)

    // ✅ USAR SERVIÇOS LIMPOS (sem *gorm.DB)
    productController := products.NewProductController(bridge.ProductService())

    // ... routes
}
```

---

## 🎯 Benefícios

| Antes | Depois |
|-------|--------|
| ❌ `*gorm.DB` em todo controller | ✅ Bridge encapsula DB |
| ❌ 10+ funções com `*gorm.DB` | ✅ 1 função: `NewBridge()` |
| ❌ Acoplamento forte | ✅ Interfaces limpas |
| ❌ Difícil de testar | ✅ Fácil mockar Bridge |

---

## 📋 Como Migrar (Gradual)

### Passo 1: Criar Bridge no router

```go
bridge := gdbasez.NewBridge(dbGorm)
```

### Passo 2: Usar métodos da Bridge

```go
// Antes
userService := gdbasez.NewUserService(gdbasez.NewUserRepo(dbGorm))

// Depois
userService := bridge.UserService()
```

### Passo 3: Remover referências a *gorm.DB

```go
// Não precisa mais passar dbGorm para controllers
controller := users.NewUserController(bridge.UserService())
```

---

## 🔄 Compatibilidade

A bridge é **100% compatível** com o código existente!

Você pode migrar **gradualmente**:

- Código antigo: continua funcionando
- Código novo: usa bridge limpa
- Sem breaking changes

---

## 🚀 Exemplo Completo: OAuth Routes

```go
func NewOAuthRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
    dbService := rtl.GetDatabaseService()
    dbGorm, _ := dbService.GetDB()

    // ✅ CRIAR BRIDGE (único ponto com *gorm.DB)
    bridge := gdbasez.NewBridge(dbGorm)

    // ✅ TUDO LIMPO daqui pra frente
    oauthService := oauthsvc.NewOAuthService(
        bridge.OAuthClientService(),
        bridge.AuthCodeService(),
        bridge.UserService(),
        tokenService,
    )

    controller := oauth.NewOAuthController(oauthService)

    // ... routes
}
```

---

## 📊 Impacto da Refatoração

### Arquivos Afetados

- ✅ `bridge.go` (novo) - Bridge centralizada
- 🔧 `gdbase_models.go` (deprecar gradualmente)
- 🔧 Routers (migrar para usar bridge)

### Zero Breaking Changes

- ✅ Código antigo continua funcionando
- ✅ Migração gradual
- ✅ Testes não quebram

---

## 🎯 Próximos Passos

1. **Usar Bridge em OAuth** ✅ (já implementado)
2. **Migrar rotas existentes** (gradual)
3. **Deprecar funções antigas** (quando todos migrarem)
4. **Remover `gdbase_models.go`** (último passo)
