# API ClasseViva usate dal progetto

Questa pagina descrive le richieste ClasseViva presenti nel client, quelle
provate fuori dal client e le superfici note che non usiamo. È un inventario
del progetto, non il catalogo completo dei servizi REST di ClasseViva.

## Attendibilità e perimetro

Le rotte `/rest/v1` qui riportate sono ricavate dal codice e da verifiche
operative del progetto. Non sono una specifica ufficiale di Spaggiari e
possono cambiare senza preavviso. Nelle fonti pubbliche ufficiali consultate
sono documentati i flussi dell'interfaccia utente, non il contratto REST;
per esempio, la [guida ufficiale alla Bacheca](https://www.spaggiari.eu/assistenza-classeviva-famiglie/assistenza-classeviva-segreterie/bacheca-consulta-la-bacheca)
descrive l'apertura di una comunicazione e il download del suo allegato.

Le risposte JSON sono restituite quasi senza trasformazioni. Il progetto non
mantiene uno schema completo dei payload restituiti dal registro. Per questo
la tabella documenta metodo, percorso e comportamento osservabile, non tutti
i campi dei dati scolastici.

## API mobile in produzione

Base URL usata dal client:

```text
https://web.spaggiari.eu/rest/v1
```

| Metodo e percorso | Strumento MCP | Richiesta e risposta | Effetti |
| --- | --- | --- | --- |
| `POST /auth/login` | Interno al client | Invia i campi `uid`, `pass`, `ident`; estrae `token` dalla risposta JSON. | Crea una sessione API, non modifica dati scolastici. |
| `GET /students/{studentId}/card` | `get_profile` | Nessun parametro; restituisce il JSON della scheda. | Lettura. |
| `GET /students/{studentId}/agenda/all/{from}/{to}` | `get_agenda` | Date `YYYYMMDD`, intervallo inclusivo; restituisce il JSON dell'agenda (compiti, eventi e lezioni). | Lettura. |
| `GET /students/{studentId}/lessons/today` | `list_lessons` | Nessun parametro; restituisce il JSON delle lezioni del giorno scelto dal servizio. | Lettura. |
| `GET /students/{studentId}/lessons/{date}` | `list_lessons` | `date` è `YYYYMMDD`; restituisce il JSON delle lezioni di quel giorno. | Lettura. |
| `GET /students/{studentId}/lessons/{from}/{to}` | `list_lessons` | Date `YYYYMMDD`, intervallo inclusivo; restituisce il JSON delle lezioni. | Lettura. |
| `GET /students/{studentId}/absences/details` | `list_absences` | Nessun parametro; restituisce il JSON dei dettagli di tutte le assenze. | Lettura. |
| `GET /students/{studentId}/absences/details/{from}` | `list_absences` | `from` è `YYYYMMDD`; restituisce i dettagli dalla data indicata. | Lettura. |
| `GET /students/{studentId}/absences/details/{from}/{to}` | `list_absences` | Date `YYYYMMDD`, intervallo inclusivo; restituisce i dettagli delle assenze. | Lettura. |
| `GET /students/{studentId}/noticeboard` | `list_noticeboard` | Nessun parametro; restituisce il JSON degli avvisi e dei metadati, inclusi gli indici degli allegati quando presenti. | Lettura; non segna gli avvisi come letti. |
| `GET /students/{studentId}/noticeboard/attach/{eventCode}/{pubId}/{attachNum}` | `get_notice_attachment` | `attachNum` va preso da `attachments[].attachNum`; nel tool `attach_num` è predefinito a `1`. Il body riuscito viene salvato come byte originali, senza convertirlo in JSON o testo in base al MIME. | Lettura remota; il tool salva il file nel percorso assoluto `destination_path` fornito dall'agente e restituisce percorso, MIME type e dimensione. Non segna l'avviso come letto e non sovrascrive file esistenti. |
| `GET /students/{studentId}/didactics` | `list_didactics` | Nessun parametro; restituisce il JSON dei materiali didattici e dei metadati. | Lettura. |
| `GET /students/{studentId}/didactics/item/{contentId}` | `download_didactics_item` | `contentId` proviene da `content_id`; il body riuscito viene trattato come byte originali del file, senza convertirlo in JSON o testo in base al MIME. | Lettura remota; il tool salva il file nel percorso assoluto `destination_path` fornito dall'agente e restituisce percorso, MIME type e dimensione. Non sovrascrive file esistenti. |

I due tool di download non sono duplicati: `get_notice_attachment` segue
`list_noticeboard` e usa gli identificativi dell'avviso più `attach_num`;
`download_didactics_item` segue `list_didactics` e usa `content_id`. Entrambi
richiedono un percorso locale assoluto, salvano il body originale e restituiscono
solo percorso, MIME type e dimensione, senza incorporare i byte nel risultato
MCP. Solo il download di un allegato della bacheca può dipendere dallo stato di
lettura dell'avviso; nessuno dei due tool modifica tale stato.

Gli identificativi `eventCode`, `pubId` e `contentId` sono richiesti e limitati
a 160 caratteri dal client; i segmenti dinamici sono codificati per uso nei
percorsi URL. `attachNum` deve essere un intero positivo. Non usare `101` come
numero allegato: quel suffisso è stato osservato sulla rotta di marcatura di
lettura, mentre i download usano i valori `attachments[].attachNum`.

`studentId` è derivato dall'UID locale. Per UID nel formato `S` + otto cifre
e una lettera finale, il client usa le otto cifre centrali; per gli altri formati
applica la conversione legacy che rimuove il primo carattere. I valori UID e
password non devono comparire negli esempi, nei log o nella documentazione.

### Autenticazione e trasporto

- Il file locale contiene UID alla prima riga e password alla seconda. La
  password è inviata solo a `POST /auth/login`; non viene restituita dagli
  strumenti MCP.
- Il corpo di login ha i campi `uid`, `pass` e `ident` (attualmente vuoto).
  Il client legge `token` dalla risposta.
- Le richieste applicative includono l'header di token `Z-Auth-Token`, un
  header tecnico `Z-DEV-ApiKey`, un `User-Agent` compatibile con il client
  mobile e `Accept: application/json, application/octet-stream, */*`. Il
  valore dell'header tecnico non va copiato in documenti o log.
- Il token resta in memoria del processo. Se una richiesta risponde `401` o
  `403`, il client invalida il token, ripete il login e riprova una volta.
- Ogni richiesta HTTP ha timeout di 30 secondi. Le risposte sono limitate a
  16 MiB.
- Il client interpreta come JSON i body JSON validi; se non sono JSON li
  tratta come file e conserva il MIME type restituito dal server, oppure lo
  rileva dai byte quando manca l'header.
- Gli errori HTTP esposti al chiamante riportano il codice di stato, non il
  body completo dell'errore. Le risposte JSON con una chiave `error` sono
  trasformate in un errore generico.

### Regole delle date

Gli strumenti accettano date ISO `YYYY-MM-DD`; il client le converte nel
percorso API in `YYYYMMDD`. Gli intervalli sono inclusivi, devono avere
`from <= to` e la distanza tra gli estremi non può superare 366 giorni.
`list_lessons` può omettere le date per usare `/lessons/today`, oppure fornire
`date` singola o `from` e `to` insieme. `list_absences` può omettere entrambe
le date, fornire solo `from` o fornire l'intervallo completo.

## API conosciute ma non esposte come strumenti MCP

### Marcatura di lettura della Bacheca

```text
POST /students/{studentId}/noticeboard/read/{eventCode}/{pubId}/101
```

Questa rotta è stata usata solo durante una verifica manuale, dopo consenso
esplicito dell'utente; non è chiamata dal codice di produzione e il server
MCP non espone un tool per invocarla. Segna l'avviso come letto e quindi
modifica lo stato su ClasseViva. Il suffisso `101` appartiene alla rotta di
marcatura e non è l'indice dell'allegato.

In una verifica reale, i GET di allegati con `attachNum` corretto hanno
risposto HTTP 404; il body della risposta diretta conteneva
`152:CvvRestApi/access to content not allowed`. Dopo un POST di lettura con
HTTP 200, gli stessi GET hanno restituito i file. La lettura era quindi
necessaria per quel caso. Questo risultato non dimostra che ogni avviso o
account abbia lo stesso vincolo. Il client MCP di produzione scarta il body
delle risposte non 2xx e restituisce soltanto un errore HTTP, per esempio
`ClasseViva request failed with HTTP 404`; non mostra al chiamante il codice
`152`. Un 404 può anche avere altre cause: prima di ipotizzare lo stato non
letto, verificare event code, publication ID e `attachNum` nei metadati. Se
resta plausibile la condizione di lettura, spiegarla come possibilità e
chiedere il consenso prima di qualsiasi marcatura. Non effettuare la
marcatura in modo automatico dentro `get_notice_attachment`.

### API web `/rest/w1`

Il client di produzione non usa `/rest/w1`: usa `/rest/v1` e l'autenticazione
mobile descritta sopra. La famiglia web richiede una sessione web separata; un
token mobile non è stato accettato in una prova di quella superficie. Una
verifica diagnostica di `whoami` ha risposto HTTP 403. Non riutilizzare il
token mobile su `/rest/w1` e non considerare questa famiglia disponibile al
client senza una nuova implementazione e verifica.

## Superfici che il progetto non usa

- Nel codice non ci sono chiamate `PUT`, `PATCH` o `DELETE` a ClasseViva.
- Il solo `POST` del client di produzione è il login. Il POST di marcatura
  della Bacheca è stato diagnostico e non è esposto dal server.
- Non sono implementate API per modificare assenze, presenze, voti, adesioni,
  risposte alle circolari o altri dati scolastici.
- Le API di altri moduli ClasseViva e gli altri endpoint non sono elencati
  perché non vengono chiamati o verificati in questo repository. Non
  dedurre percorsi o capacità dal nome di un modulo. Questa pagina non è una
  catalogazione esaustiva del backend Spaggiari.

## Endpoint locali esclusi dalla API ClasseViva

La procedura `begin_credentials_setup` avvia un server HTTP temporaneo del
progetto su `127.0.0.1` con un link casuale sotto `/setup/{token}`:

- `GET /setup/{token}` mostra il modulo locale di configurazione.
- `POST /setup/{token}` salva UID e password nella configurazione privata
  locale; non inoltra il modulo all'API ClasseViva.

Il token è monouso, il link scade dopo 10 minuti e il server accetta richieste
solo da loopback con host e origine attesi. Questi due endpoint appartengono
all'interfaccia locale del plugin, non al servizio ClasseViva.

## Dove mantenere aggiornato l'inventario

- Implementazione API mobile: `internal/classeviva/client.go`.
- Associazione tra strumenti MCP e metodi client: `cmd/classeviva/main.go`.
- Salvataggio locale dei materiali: `cmd/classeviva/download.go` e
  `internal/classeviva/download.go`.
- Pagina locale delle credenziali: `internal/setupui/setup.go`.

Quando si aggiunge, rimuove o cambia una rotta, aggiornare questa pagina con
metodo, percorso, parametri, effetto, tool associato e prova disponibile.
Separare sempre le chiamate di produzione dalle prove manuali e dalle rotte
soltanto ipotizzate.

## Copertura delle verifiche nel repository

I test unitari verificano in particolare la costruzione dei percorsi per le
varianti di `list_absences` e per gli allegati 1 e 2, oltre al rifiuto di
indici non positivi. Sono test con trasporto HTTP simulato: non attestano che
ClasseViva mantenga quelle rotte disponibili. La prova live degli allegati e
della marcatura di lettura è stata limitata a un avviso reale, senza includere
nel repository il titolo, l'identificativo o i contenuti personali.
