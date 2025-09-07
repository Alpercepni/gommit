# gommit
[![pt-BR](https://img.shields.io/badge/🇧🇷-Português-black)](README.pt-BR.md)
[![ru](https://img.shields.io/badge/🇷🇺-Русский-black)](README.ru.md)
[![hi](https://img.shields.io/badge/🇮🇳-Hindi-black)](README.hi.md)
[![zh-CN](https://img.shields.io/badge/🇨🇳-中文-black)](README.zh-CN.md)
[![es](https://img.shields.io/badge/🇪🇸-Español-green)](README.es.md)
[![en](https://img.shields.io/badge/🇺🇸-English-black)](README.md)

**gommit** es un asistente de línea de comandos rápido y sin dependencias para _Conventional Commits_, escrito en Go.  
Abre un asistente interactivo (similar a Commitizen/cz) y ejecuta `git commit` con el mensaje correctamente formateado.

> 💡 Por defecto, **emojis de tipo de commit** se incluyen **en el encabezado** (ej: `feat 💡: ...`) para hacerlos visibles en las listas de archivos/carpetas de GitHub. Los emojis solo se añaden al título, no al cuerpo o pie de página.

---

## ✨ Características

- ✅ Asistente interactivo con búsqueda/atajos (números, palabras clave, `q` para salir)
- ✅ Tipos estándar de Conventional Commits: **feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert**  
  Más extras: **WIP, prune**
- ✅ Emojis en el menú (con fallback ASCII automático) y **emoji en el encabezado del commit**
- ✅ Verificaciones previas de Git:
  - Validación de repositorio (`not a git repository…` si no está en uno)
  - Auto-stage (`git add -A`) cuando no hay nada en stage (configurable)
  - Lista cambios en stage antes de hacer commit (configurable)
  - Modo `--amend` con banner del último commit
- ✅ Campos completos estilo Commitizen:
  - `scope` (opcional)
  - `subject` (requerido, `<=72` caracteres)
  - `body` multilínea (termina con **punto único en una línea** **o** **presiona Enter dos veces**)
  - `BREAKING CHANGE` (pregunta y pide descripción)
  - Issues: `Closes #123` / `Refs #45` (si se desea)
- ✅ Modo **`--install`** integrado que instala/actualiza el binario en el PATH del usuario
  - Windows: `%LOCALAPPDATA%\Programs\gommit\bin`
  - Linux/macOS: `~/.local/bin` (añade al PATH si falta)

---

## 🚀 Instalación

### Método 1: Scripts de Instalación de Una Línea **(Recomendado)**

**Windows:**
```powershell
powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"
```

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
```

### Método 2: Descargar Binario de Release
1. Descarga el archivo para tu OS/arquitectura desde la página de **Releases**
2. Extrae el paquete (`.zip` en Windows / `.tar.gz` en Linux/macOS)
3. Ejecuta el binario extraído con la flag `--install`:

**Windows (PowerShell):**
```powershell
.\gommit.exe --install
# Cierra y reabre la terminal
gommit --version
```

**Linux/macOS:**
```bash
chmod +x ./gommit
./gommit --install
# Reinicia el shell
gommit --version
```

> Ejecutar `--install` nuevamente actualizará la instalación.

---

## 💻 Uso Rápido

Dentro de un repositorio Git con cambios:

```bash
gommit
```

Flujo típico del asistente:
1. Selecciona **tipo** (número, nombre o búsqueda)
2. Ingresa **scope** (opcional)
3. Escribe **subject** (modo imperativo, `<=72` chars)
4. Añade **body** (opcional, multilínea) – termina con `.` en una línea **o** presiona Enter **dos veces**
5. **¿Cambios breaking?** (si "sí", describe el cambio)
6. **¿Issues?** (Closes/Refs)

Al final, `gommit` ejecuta `git commit` (o muestra preview con `--dry-run`).

### Ejemplos

Commit normal (con auto-stage):
```bash
gommit
```

Permitir commit vacío:
```bash
gommit --allow-empty
```

Enmendar último commit:
```bash
gommit --amend
```

Mostrar solo el mensaje (no hacer commit):
```bash
gommit --dry-run
```

Pasar todo vía flags (sin asistente):
```bash
gommit --type feat --scope ui --subject "añadir selector de tipo" --body "Detalles de implementación..."
```

Modo no interactivo con todos los parámetros:
```bash
gommit --type fix --scope api --subject "resolver problema de autenticación" --body "Arreglada validación de token JWT\n\nEsto resuelve el problema donde tokens expirados\nno se manejaban adecuadamente." --footer "Closes #42"
```

---

## 📝 Formato del Encabezado del Commit

```
<type>[(!)][(scope)] <emoji>: <subject>
```

Ejemplos:
```
feat 💡: añadir comando install
fix(api) 🐛: corregir nil pointer en carga de config
refactor(core)! 🎨: unificar message builder
```

> El `!` aparece solo en el **encabezado**; la validación interna usa el tipo "puro" (`feat`, `fix`, etc.) para mantener compatibilidad con herramientas de Conventional Commits.

---

## ⚙️ Flags de Línea de Comandos

| Flag | Descripción |
|---|---|
| `--version` | Muestra versión y sale |
| `--install` | Instala/actualiza binario en el PATH del usuario y sale |
| `--dry-run` | Solo imprime el mensaje generado (no llama a `git commit`) |
| `--type` | Tipo de commit (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert, WIP, prune) |
| `--scope` | Scope opcional (ej: `ui`, `api`) |
| `--subject` | Línea de asunto (modo imperativo, `<=72` chars) |
| `--body` | Texto del cuerpo (usa `\n` para nuevas líneas) |
| `--footer` | Pie de página manual (ej: `Closes #123`) |
| `--as-editor` | Ejecutar como editor de Git (modo `COMMIT_EDITMSG`) |
| `--allow-empty` | Permitir commit vacío |
| `--amend` | Enmendar el último commit |
| `--no-verify` | Saltar hooks (`pre-commit`, `commit-msg`) |
| `--signoff` | Añadir trailer `Signed-off-by` |
| `--auto-stage` | Auto `git add -A` cuando no hay nada en stage (por defecto: `true`) |
| `--show-status` | Mostrar resumen de cambios en stage antes del commit (por defecto: `true`) |

---

## 📋 Tipos de Commit Soportados

| Tipo | Emoji | Descripción |
|---|---|---|
| `WIP` | 🚧 | Trabajo en progreso |
| `feat` | 💡 | Una nueva funcionalidad |
| `fix` | 🐛 | Arreglar un bug |
| `chore` | 📦 | Actualizar dependencias, despliegues, archivos de configuración |
| `refactor` | 🎨 | Mejorar estructura/formato del código |
| `prune` | 🔥 | Eliminar código o archivos |
| `docs` | 📝 | Escribir documentación |
| `perf` | ⚡ | Mejorar rendimiento |
| `test` | ✅ | Añadir tests |
| `build` | 🔧 | Cambios en el sistema de build o dependencias |
| `ci` | 🤖 | Cambios en configuración de CI/CD |
| `style` | 💅 | Cambios que no afectan el significado del código |
| `revert` | ⏪ | Revertir a un commit |

---

## 🔧 Modo Editor de Git

Puedes configurar `gommit` como tu editor de Git para mensajes de commit consistentes:

```bash
git config --global core.editor "gommit --as-editor"
```

Esto abrirá el asistente `gommit` cuando Git necesite un mensaje de commit (incluyendo durante `git commit`, `git merge`, `git rebase`, etc.).

---

## 🛠️ Desarrollo

### Construir desde el Código Fuente

Requisitos: Go 1.22+

```bash
# Clonar el repositorio
git clone https://github.com/Hangell/gommit.git
cd gommit

# Construir
go build -o gommit ./cmd/gommit

# Ejecutar tests
go test ./...

# Instalar localmente
./gommit --install
```

### Construir con Información de Versión

Para releases con inyección de versión:
```bash
go build -ldflags "-s -w -X main.version=0.2.0" -o gommit ./cmd/gommit
```

---

## 🎨 Variables de Entorno

| Variable | Descripción |
|---|---|
| `NO_COLOR=1` | Deshabilita colores ANSI en el menú |
| `NO_EMOJI=1` | Fuerza fallback ASCII en el menú (encabezado aún usa emojis unicode) |

---

## 🤝 Contribuyendo

1. Haz fork del repositorio
2. Crea tu rama de feature (`git checkout -b feature/caracteristica-increible`)
3. Haz commit de tus cambios usando `gommit` 😉
4. Haz push a la rama (`git push origin feature/caracteristica-increible`)
5. Abre un Pull Request

---

## 📄 Licencia

GPL-3.0-only — ver archivo [`LICENSE`](LICENSE) para detalles.

---

## ❓ FAQ

**¿gommit reemplaza a Commitizen?** Es una **alternativa**. Si tu entorno ya depende de Node/npm y estás cómodo con adaptadores cz, quédate con él. Si quieres **menos dependencias**, **instalación única**, y **uso consistente** en múltiples repositorios/servidores, gommit es probablemente más simple.

**¿Es más rápido?** No publicamos benchmarks aquí, pero como binario nativo, gommit **tiende a iniciar más rápido** y usar menos memoria que herramientas CLI tradicionales de Node — especialmente en entornos fríos (CI, containers) donde iniciar el runtime de Node añade latencia.

**¿Funciona offline?** Sí. Una vez instalado, no necesita acceso a internet para ejecutar el asistente o hacer commit.

**¿Por qué se creó gommit?** La idea nació de la frustración con entornos nvm donde teníamos que instalar `git-cz` para cada versión de nvm o directamente en cada proyecto. gommit es mucho más simple y proporciona una solución única y unificada.

---

## 👨‍💻 Autor

<div align="center">
  <img src="https://avatars.githubusercontent.com/u/53544561?v=4" width="150" style="border-radius: 50%;" />

**Rodrigo Rangel**

  <div>
    <a href="https://hangell.org" target="_blank">
      <img src="https://img.shields.io/badge/website-000000?style=for-the-badge&logo=About.me&logoColor=white" alt="Website" />
    </a>
    <a href="https://play.google.com/store/apps/dev?id=5606456325281613718" target="_blank">
      <img src="https://img.shields.io/badge/Google_Play-414141?style=for-the-badge&logo=google-play&logoColor=white" alt="Google Play" />
    </a>
    <a href="https://www.youtube.com/channel/UC8_zG7RFM2aMhI-p-6zmixw" target="_blank">
      <img src="https://img.shields.io/badge/YouTube-FF0000?style=for-the-badge&logo=youtube&logoColor=white" alt="YouTube" />
    </a>
    <a href="https://www.facebook.com/hangell.org" target="_blank">
      <img src="https://img.shields.io/badge/Facebook-1877F2?style=for-the-badge&logo=facebook&logoColor=white" alt="Facebook" />
    </a>
    <a href="https://www.linkedin.com/in/rodrigo-rangel-a80810170" target="_blank">
      <img src="https://img.shields.io/badge/-LinkedIn-%230077B5?style=for-the-badge&logo=linkedin&logoColor=white" alt="LinkedIn" />
    </a>
  </div>
</div>

---

## 🙏 Agradecimientos

- Inspirado por [Commitizen](https://github.com/commitizen/cz-cli)
- Construido con la especificación [Conventional Commits](https://www.conventionalcommits.org/)
- ¡Los emojis ayudan a hacer el historial de commits más visual y divertido! 🎉

---

<div align="center">
  <strong>Hecho con ❤️ para la comunidad de desarrolladores</strong>
</div>