# Chirpy - Twitter-like REST API

API REST construida en Go para crear y gestionar "chirps" (mensajes cortos similares a tweets) con autenticación JWT.

## Tecnologías

- **Go 1.24.3**
- **PostgreSQL** con migraciones Goose
- **SQLc** para consultas SQL type-safe
- **JWT** para autenticación (golang-jwt/jwt/v5)
- **Argon2id** para hashing de contraseñas
- **net/http** (sin frameworks externos)

## Estructura del Proyecto

```
chirpy/
├── main.go                 # Entry point y configuración del servidor
├── handlers.go             # Manejadores HTTP (293 líneas)
├── types.go                # Modelos de dominio
├── utils.go                # Utilidades y helpers
├── internal/
│   ├── auth/              # Lógica de autenticación (JWT, passwords)
│   └── database/          # Código generado por SQLc
├── sql/
│   ├── schema/            # Migraciones de base de datos (Goose)
│   └── queries/           # Consultas SQL para SQLc
└── testing/
    └── auth_test.go       # Tests unitarios de autenticación
```

## Endpoints Disponibles

### API Endpoints
| Método | Ruta | Descripción | Autenticación |
|--------|------|-------------|---------------|
| GET | `/api/healthz` | Health check | No |
| GET | `/api/chirps` | Obtener todos los chirps | No |
| GET | `/api/chirps/{chirpId}` | Obtener chirp por ID | No |
| POST | `/api/users` | Crear usuario | No |
| POST | `/api/login` | Login de usuario | No |
| POST | `/api/chirps` | Crear chirp | Sí (JWT) |
| POST | `/api/refresh` | Refrescar token de acceso | Sí (Refresh Token) |
| POST | `/api/revoke` | Revocar refresh token | Sí (Refresh Token) |

### Admin Endpoints (Solo desarrollo)
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/admin/metrics` | Ver contador de visitas |
| POST | `/admin/reset` | Resetear base de datos |

---

## Áreas de Mejora y Refactorización

### Prioridad 1: Crítico ⚠️

#### 1. **Manejo de Errores en Decodificación JSON**
**Problema:** Se ignoran errores al decodificar JSON, lo que puede causar procesamiento de datos vacíos.

**Ubicación:**
- [handlers.go:74](handlers.go#L74) - `usersHandler` ✅
- [handlers.go:117](handlers.go#L117) - `chirpsHandler` ✅
- [handlers.go:185](handlers.go#L185) - `loginHandler` ✅

**Código actual:**
```go
decoder.Decode(&userReqdata)  // Error ignorado
```

**Solución recomendada:**
```go
if err := decoder.Decode(&userReqdata); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
    return
}
```

#### 2. **Panic Potencial con UUID** ✅
**Problema:** Uso de `uuid.MustParse()` que causa panic si el UUID es inválido.

**Ubicación:** [handlers.go:167](handlers.go#L167)

**Código actual:**
```go
dbChirp, err := ac.Queries.GetChirpById(r.Context(), uuid.MustParse(chirpId))
```

**Solución recomendada:**
```go
chirpUUID, err := uuid.Parse(chirpId)
if err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format")
    return
}
dbChirp, err := ac.Queries.GetChirpById(r.Context(), chirpUUID)
```

#### 3. **Límite de Tamaño de Request** ✅
**Problema:** No hay límites en el tamaño del body, permitiendo potenciales ataques DoS.

**Solución recomendada:**
```go
// En cada handler antes de decodificar:
r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB límite
decoder := json.NewDecoder(r.Body)
```

#### 4. **Typo en Nombre de Campo** ✅
**Problema:** `TokenScret` debería ser `TokenSecret` en toda la codebase.

**Ubicación:**
- [types.go](types.go) - Definición del struct
- [main.go:39](main.go#L39)
- [handlers.go:125](handlers.go#L125)
- [handlers.go:205](handlers.go#L205)
- [handlers.go:266](handlers.go#L266)

**Impacto:** Aunque funcional, afecta la profesionalidad del código.

---

### Prioridad 2: Alta 🔴

#### 5. **Implementar Rate Limiting** ✅
**Problema:** Sin protección contra ataques de fuerza bruta en el endpoint de login.

**Solución recomendada:**
```go
// Usar github.com/didip/tollbooth o similar
import "github.com/didip/tollbooth/v7"

limiter := tollbooth.NewLimiter(5, nil) // 5 requests por segundo
mux.Handle("POST /api/login", tollbooth.LimitFuncHandler(limiter, apiCfg.loginHandler))
```

#### 6. **Agregar Logging de Requests** ✅
**Problema:** No hay registro de peticiones HTTP, dificultando el debugging en producción.

**Solución recomendada:**
```go
// Middleware de logging
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
        next.ServeHTTP(w, r)
        log.Printf("Completed in %v", time.Since(start))
    })
}

// Aplicar a todas las rutas:
s := &http.Server{
    Addr:    ":" + PORT,
    Handler: loggingMiddleware(mux),
}
```

#### 7. **Paginación en GET /api/chirps**
**Problema:** El endpoint retorna TODOS los chirps sin paginación.

**Ubicación:** [handlers.go:150-163](handlers.go#L150-L163)

**Solución recomendada:**
```go
// Agregar parámetros de query
func (ac *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
    page := r.URL.Query().Get("page")
    limit := r.URL.Query().Get("limit")

    // Valores por defecto
    pageNum := 1
    limitNum := 50

    // Parsear y validar...
    // Modificar query SQL para incluir LIMIT y OFFSET
}
```

#### 8. **Validación de Contraseñas**
**Problema:** No hay requisitos mínimos de seguridad para contraseñas.

**Ubicación:** [handlers.go:76-79](handlers.go#L76-L79)

**Solución recomendada:**
```go
func validatePassword(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    // Agregar validaciones adicionales:
    // - Al menos una mayúscula
    // - Al menos un número
    // - Al menos un carácter especial
    return nil
}
```

#### 9. **Validación de Email**
**Problema:** No se valida el formato del email.

**Solución recomendada:**
```go
import "net/mail"

func validateEmail(email string) error {
    _, err := mail.ParseAddress(email)
    if err != nil {
        return errors.New("invalid email format")
    }
    return nil
}
```

---

### Prioridad 3: Media 🟡

#### 10. **Extraer Autenticación a Middleware**
**Problema:** La extracción del Bearer token se repite en múltiples handlers.

**Ubicación:**
- [handlers.go:119](handlers.go#L119) - `chirpsHandler`
- [handlers.go:238](handlers.go#L238) - `refreshTokenHandler`
- [handlers.go:280](handlers.go#L280) - `revokeTokenHandler`

**Solución recomendada:**
```go
// Middleware de autenticación
func (ac *apiConfig) requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token, err := auth.GetBearerToken(r.Header)
        if err != nil {
            respondWithError(w, http.StatusUnauthorized, "Unauthorized")
            return
        }

        userID, err := auth.ValidateJWT(token, ac.TokenSecret)
        if err != nil {
            respondWithError(w, http.StatusUnauthorized, "Invalid token")
            return
        }

        // Añadir userID al contexto
        ctx := context.WithValue(r.Context(), "userID", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

// Uso:
mux.HandleFunc("POST /api/chirps", ac.requireAuth(ac.chirpsHandler))
```

#### 11. **Palabras Prohibidas Configurables**
**Problema:** Las palabras prohibidas están hardcodeadas.

**Ubicación:** [handlers.go:136](handlers.go#L136)

**Solución recomendada:**
```go
// En apiConfig
type apiConfig struct {
    fileserverhits atomic.Int32
    Queries        *database.Queries
    Env            string
    TokenSecret    string
    BannedWords    []string  // Nueva configuración
}

// Cargar desde archivo o env variable
bannedWords := strings.Split(os.Getenv("BANNED_WORDS"), ",")
```

#### 12. **Validación de Content-Type Inconsistente**
**Problema:** Solo `chirpsHandler` valida Content-Type.

**Ubicación:** [handlers.go:108](handlers.go#L108)

**Solución recomendada:**
```go
// Middleware para validar Content-Type en POSTs
func requireJSON(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" && r.Header.Get("Content-Type") != "application/json" {
            respondWithError(w, http.StatusBadRequest, "Content-Type must be application/json")
            return
        }
        next.ServeHTTP(w, r)
    }
}
```

#### 13. **Gestión de Múltiples Refresh Tokens**
**Problema:** Cada login crea un nuevo refresh token sin limpiar los antiguos.

**Comentario en código:** [handlers.go:176-179](handlers.go#L176-L179)

**Solución recomendada:**
```go
// Antes de crear nuevo token, revocar tokens existentes del usuario
func (ac *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
    // ... validaciones ...

    // Revocar tokens anteriores del usuario
    err = ac.Queries.RevokeUserTokens(r.Context(), dbUser.ID)
    if err != nil {
        log.Printf("Error revoking old tokens: %v", err)
        // Continuar aunque falle
    }

    // Crear nuevo token...
}
```

Requiere nueva query SQL:
```sql
-- name: RevokeUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;
```

#### 14. **Mensajes de Error Genéricos**
**Problema:** Mensajes como "Something went wrong" no son informativos.

**Solución recomendada:**
```go
// Crear constantes para errores comunes
const (
    ErrInvalidJSON      = "Invalid JSON format"
    ErrChirpTooLong     = "Chirp must be 140 characters or less"
    ErrInvalidChirpID   = "Invalid chirp ID"
    ErrUnauthorized     = "Unauthorized"
    ErrInvalidCredentials = "Invalid email or password"
)
```

#### 15. **Configuración de CORS**
**Problema:** No hay headers CORS configurados.

**Solución recomendada:**
```go
// Middleware CORS
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

### Prioridad 4: Baja 🟢

#### 16. **Expandir Cobertura de Tests**
**Problema:** Solo hay 4 tests unitarios en [testing/auth_test.go](testing/auth_test.go).

**Tests faltantes:**
- Tests de handlers HTTP
- Tests de integración con base de datos
- Tests de edge cases (tokens expirados, usuarios duplicados, etc.)

**Ejemplo de test de handler:**
```go
func TestChirpsHandler(t *testing.T) {
    // Mock database
    // Mock request con JWT válido
    // Verificar respuesta
}
```

#### 17. **Validación de Variables de Entorno**
**Problema:** No se valida que las variables de entorno existan al inicio.

**Ubicación:** [main.go:23-25](main.go#L23-L25)

**Solución recomendada:**
```go
func validateEnv() error {
    required := []string{"DB_URL", "TOKEN_SECRET"}
    for _, key := range required {
        if os.Getenv(key) == "" {
            return fmt.Errorf("required environment variable %s is not set", key)
        }
    }
    return nil
}

func main() {
    godotenv.Load()

    if err := validateEnv(); err != nil {
        log.Fatal(err)
    }

    // ...
}
```

#### 18. **Métricas Mejoradas**
**Problema:** Solo se cuenta visitas a `/app/*`, no persistente, no en JSON.

**Solución recomendada:**
- Usar Prometheus para métricas
- Almacenar métricas en Redis
- Endpoint JSON en lugar de HTML

#### 19. **Graceful Shutdown**
**Problema:** El servidor no maneja señales de terminación correctamente.

**Solución recomendada:**
```go
func main() {
    // ... setup ...

    // Canal para señales
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    go func() {
        log.Fatal(s.ListenAndServe())
    }()

    <-stop

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := s.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }

    log.Println("Server stopped gracefully")
}
```

#### 20. **Estructura de Directorios Mejorada**
**Recomendación:** Adoptar estructura más estándar de Go:

```
chirpy/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   └── models/
├── pkg/           # Código reutilizable
├── migrations/
└── tests/
```

#### 21. **Documentación de API**
**Problema:** No hay documentación OpenAPI/Swagger.

**Solución recomendada:**
- Usar `swaggo/swag` para generar docs automáticamente
- Añadir comentarios en handlers para generar OpenAPI spec

#### 22. **Connection Pooling Explícito**
**Problema:** Se usa configuración por defecto del pool de conexiones.

**Solución recomendada:**
```go
db, err := sql.Open("postgres", dbUrl)
if err != nil {
    log.Fatal("Error setting up the database")
}

db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Ping para verificar conexión
if err := db.Ping(); err != nil {
    log.Fatal("Cannot connect to database:", err)
}
```

---

## Mejoras de Seguridad

### Implementadas ✅
- Argon2id para hashing de contraseñas
- JWT con HS256
- Refresh token rotation
- Revocación de tokens
- Foreign key constraints
- UUIDs en lugar de IDs incrementales

### Pendientes ⚠️
- Rate limiting (fuerza bruta)
- HTTPS enforcement
- Límites de tamaño de request
- CORS configurado correctamente
- Logging de eventos de seguridad
- Password strength requirements
- Email verification
- 2FA (opcional)

---

## Refactorizaciones Arquitectónicas Sugeridas

### 1. **Separar Handlers en Archivos**
```
internal/handlers/
├── health.go
├── users.go
├── chirps.go
├── auth.go
└── admin.go
```

### 2. **Capa de Servicio**
Separar lógica de negocio de handlers:

```go
// internal/service/chirp_service.go
type ChirpService struct {
    queries *database.Queries
}

func (s *ChirpService) CreateChirp(ctx context.Context, userID uuid.UUID, body string) (*Chirp, error) {
    // Validación
    // Lógica de negocio
    // Llamada a DB
}
```

### 3. **Repository Pattern**
Encapsular acceso a datos:

```go
type ChirpRepository interface {
    Create(ctx context.Context, chirp *Chirp) error
    GetByID(ctx context.Context, id uuid.UUID) (*Chirp, error)
    List(ctx context.Context, opts ListOptions) ([]*Chirp, error)
}
```

---

## Comandos Útiles

### Setup
```bash
# Instalar dependencias
go mod download

# Setup base de datos
goose -dir sql/schema postgres "YOUR_DB_URL" up

# Generar código SQLc
sqlc generate
```

### Desarrollo
```bash
# Ejecutar servidor
go run .

# Ejecutar tests
go test ./...

# Ejecutar tests con coverage
go test -cover ./...
```

### Migraciones
```bash
# Crear nueva migración
goose -dir sql/schema create add_feature sql

# Ejecutar migraciones
goose -dir sql/schema postgres "DB_URL" up

# Rollback última migración
goose -dir sql/schema postgres "DB_URL" down
```

---

## Variables de Entorno Requeridas

Crear archivo `.env` en la raíz:

```env
DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
TOKEN_SECRET=tu-secreto-muy-seguro-aqui
ENV=dev
```

---

## Notas Adicionales

### Performance
- El código actual está bien optimizado para cargas pequeñas-medianas
- Para escalar, considerar:
  - Cache con Redis para chirps populares
  - Read replicas para PostgreSQL
  - CDN para assets estáticos

### Observabilidad
Añadir:
- Structured logging (zerolog, zap)
- Tracing (OpenTelemetry)
- Métricas (Prometheus)
- Health checks detallados

### CI/CD
Configurar:
- GitHub Actions para tests automáticos
- Linting con golangci-lint
- Security scanning con gosec
- Dependency updates con Dependabot

---

## Recursos

- [Go Best Practices](https://golang.org/doc/effective_go)
- [SQLc Documentation](https://docs.sqlc.dev)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Go Security](https://owasp.org/www-project-go-secure-coding-practices-guide/)

---

## Conclusión

Este proyecto tiene una base sólida con buenas prácticas de seguridad (Argon2id, JWT, refresh tokens). Las principales áreas de mejora son:

1. **Manejo de errores** (crítico)
2. **Rate limiting y logging** (seguridad)
3. **Paginación y validación** (robustez)
4. **Testing y documentación** (mantenibilidad)

Con estas mejoras, el proyecto estaría listo para producción.
