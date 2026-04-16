# 🗄️ Zaawansowane Technologie Baz Danych - Projekt

Projekt porównawczy wydajności 4 systemów zarządzania bazami danych:
- **MySQL 8.3** (relacyjna)
- **PostgreSQL 16** (relacyjna)  
- **MongoDB 7** (nierelacyjna - dokumentowa)
- **Redis 7** (nierelacyjna - klucz-wartość)

## 📋 Spis treści

- [Wymagania](#wymagania)
- [Szybki start](#szybki-start)
- [Struktura projektu](#struktura-projektu)
- [Jak to działa](#jak-to-działa)
- [Uruchamianie testów](#uruchamianie-testów)
- [Wyniki](#wyniki)
- [Scenariusze testowe](#scenariusze-testowe)

## 🎯 Wymagania

- **Go** 1.23+ ([instalacja](https://go.dev/doc/install))
- **Docker** & **Docker Compose** ([instalacja](https://docs.docker.com/get-docker/))
- **Git**
- Co najmniej 8GB RAM (dla dużych zbiorów danych)

## 🚀 Szybki start

### 1. Sklonuj repozytorium i uruchom bazy danych

```bash
# Uruchom wszystkie bazy danych w Dockerze
docker-compose up -d

# Sprawdź czy wszystko działa
docker-compose ps
```

Powinny działać 4 kontenery:
- `bench-mysql` (port 3306)
- `bench-postgres` (port 5432)
- `bench-mongo` (port 27017)
- `bench-redis` (port 6379)

### 2. Wygeneruj dane testowe

```bash
cd src
go run main.go
cd ..
```

To wygeneruje ~4GB plików z zapytaniami w katalogu `queries/`.

### 3. Uruchom szybki test

```bash
./quick_test.sh
```

To uruchomi szybki test (1000 zapytań, 1 próba) żeby sprawdzić czy wszystko działa.

### 4. Uruchom pełny benchmark

```bash
./run_benchmarks.sh --trials 3 --batch 10000
```

## 📁 Struktura projektu

```
.
├── docker-compose.yml          # Konfiguracja baz danych
├── init/                       # Skrypty inicjalizacyjne
│   ├── sql/
│   │   ├── schema.sql         # Schema dla PostgreSQL
│   │   └── schema_mysql.sql   # Schema dla MySQL
│   └── mongo/init.js          # Inicjalizacja MongoDB
├── src/
│   ├── main.go                # Generator danych testowych
│   ├── generator/             # Logika generowania zapytań
│   ├── benchrunner/           # Biblioteka do benchmarkingu
│   │   ├── config.go
│   │   ├── result.go
│   │   ├── mysql_runner.go
│   │   └── postgres_runner.go
│   └── cmd/run_benchmarks/    # Główny program benchmarkowy
├── bin/
│   └── bench                  # Skompilowany program (gitignored)
├── queries/                    # Wygenerowane zapytania (gitignored)
├── results/                    # Wyniki benchmarków (gitignored)
├── run_benchmarks.sh          # Skrypt do pełnych testów
└── quick_test.sh              # Szybki test
```

## 🔧 Jak to działa

### Faza 1: Generowanie danych

Program `src/main.go` generuje:
- **10 tabel/kolekcji**: users, products, orders, addresses, categories, etc.
- **Różne rozmiary**: 200 kategorii, 100k produktów, 1M użytkowników
- **Wszystkie operacje CRUD**: INSERT, SELECT, UPDATE, DELETE
- **Dla wszystkich baz**: SQL, MongoDB (JS), Redis (commands)

### Faza 2: Uruchamianie testów

Program `src/cmd/run_benchmarks/main.go`:
1. Łączy się z bazą danych
2. Wczytuje plik z zapytaniami
3. Wykonuje N zapytań i mierzy czas
4. Powtarza test 3 razy (trials)
5. Oblicza średnią, min, max
6. Zapisuje wyniki do JSON i CSV

### Faza 3: Analiza wyników

Wyniki zawierają:
- Całkowity czas wykonania
- Średni czas na zapytanie
- Queries per second (QPS)
- Statystyki dla każdej próby

## 🎮 Uruchamianie testów

### Podstawowe użycie

```bash
# Wszystkie bazy, 3 próby, 10k zapytań
./run_benchmarks.sh

# Tylko MySQL, 5 prób, 100k zapytań
./run_benchmarks.sh --db mysql --trials 5 --batch 100000

# Tylko PostgreSQL, szybki test
./run_benchmarks.sh --db postgres --trials 1 --batch 1000
```

### Ręczne uruchomienie

```bash
# Bezpośrednio z binarki
./bin/bench -db mysql -trials 3 -batch 10000

# Lub kompiluj i uruchom
cd src/cmd/run_benchmarks
go run main.go -db mysql -trials 3 -batch 10000
```

### Parametry

- `--db` / `-db`: Która baza testować
  - `mysql` - tylko MySQL
  - `postgres` - tylko PostgreSQL
  - `mongodb` - tylko MongoDB (TODO)
  - `redis` - tylko Redis (TODO)
  - `all` - wszystkie (domyślnie)

- `--trials` / `-trials`: Ile prób dla każdego testu (domyślnie: 3)

- `--batch` / `-batch`: Ile zapytań uruchomić (domyślnie: 10000)

## 📊 Wyniki

Wyniki są zapisywane w katalogu `results/` z timestampem:

### Format JSON

```json
{
  "results": [
    {
      "database": "mysql",
      "entity": "users",
      "operation": "insert",
      "batch_size": 10000,
      "trial": 1,
      "num_queries": 10000,
      "total_time_ms": 5234.56,
      "avg_time_per_op_ms": 0.52,
      "queries_per_second": 1910.45,
      "timestamp": "2024-03-19T10:00:00Z"
    }
  ]
}
```

### Format CSV

Łatwo otworzyć w Excel/Google Sheets:

```csv
database,entity,operation,batch_size,trial,num_queries,total_time_ms,avg_time_per_op_ms,queries_per_second
mysql,users,insert,10000,1,10000,5234.56,0.52,1910.45
mysql,users,insert,10000,2,10000,5189.23,0.52,1927.34
```

## 🧪 Scenariusze testowe

### Obecnie zaimplementowane

| Baza | Entity | Operacje | Liczba zapytań |
|------|--------|----------|----------------|
| MySQL | users | INSERT, SELECT, UPDATE | 1,000,000 |
| MySQL | products | INSERT, SELECT, UPDATE | 100,000 |
| MySQL | orders | INSERT, SELECT | 1,000,000 |
| PostgreSQL | users | INSERT, SELECT, UPDATE | 1,000,000 |
| PostgreSQL | products | INSERT, SELECT, UPDATE | 100,000 |

### TODO - Planowane rozszerzenia

- [ ] MongoDB benchmarks
- [ ] Redis benchmarks
- [ ] Zapytania złożone (JOIN, agregacje)
- [ ] Testy z indeksami vs bez indeksów
- [ ] Bulk operations
- [ ] Transakcje
- [ ] Wizualizacja wyników (wykresy)
- [ ] Automatyczny raport PDF

## 🔍 Rozwiązywanie problemów

### Bazy danych nie startują

```bash
# Sprawdź logi
docker-compose logs mysql
docker-compose logs postgres

# Restart
docker-compose down
docker-compose up -d
```

### Błąd "connection refused"

Poczekaj chwilę - bazy potrzebują czasu na inicjalizację:

```bash
docker-compose up -d
sleep 10
```

### Brak plików z zapytaniami

```bash
cd src
./benchmark
```

### Błędy kompilacji Go

```bash
cd src
go mod tidy
go mod download
```

## 📈 Następne kroki

1. **Uruchom pełny benchmark** z różnymi rozmiarami danych
2. **Dodaj MongoDB i Redis** runners
3. **Stwórz wizualizacje** wyników (Python + matplotlib)
4. **Napisz sprawozdanie** z analizą wyników
5. **Przygotuj prezentację**

## 📝 Autorzy

Projekt na zajęcia "Zaawansowane Technologie Baz Danych"

## 📄 Licencja

MIT

