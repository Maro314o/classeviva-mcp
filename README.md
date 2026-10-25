# ClasseViva

Plugin locale per Codex: un server MCP scritto in Go e la skill **ClasseViva** sono distribuiti insieme. La skill guida l'agente nell'uso naturale degli strumenti e nella configurazione iniziale. Per esempio, dopo aver installato il plugin puoi chiedere: «Che compiti ho domani?» senza menzionare né il plugin né la skill.

## Installazione da GitHub in Codex

Il repository include un marketplace Codex in `.agents/plugins/marketplace.json`. Dopo aver pubblicato il repository su GitHub, sostituisci `OWNER/REPOSITORY` con il proprietario e il nome effettivi del repository e installa il plugin dal terminale:

```sh
codex plugin marketplace add OWNER/REPOSITORY
codex plugin add classeviva@classeviva-repo
codex plugin list
```

La prima riga aggiunge il marketplace del repository; la seconda installa **ClasseViva**. La skill è inclusa nel plugin e viene resa disponibile insieme agli strumenti MCP: non serve installarla separatamente. Se Codex non mostra subito gli strumenti, apri una nuova conversazione o riavvia l'applicazione.

### Istruzioni per l'agente che esegue l'installazione

Quando l'utente chiede di installare ClasseViva seguendo questo README:

1. Usa il marketplace GitHub configurato sopra, sostituendo `OWNER/REPOSITORY` con l'indirizzo indicato dall'utente o nel contesto del repository.
2. Esegui i comandi `codex plugin marketplace add` e `codex plugin add` riportati sopra, poi controlla `codex plugin list`.
3. Se il plugin non compare nella conversazione corrente, chiedi all'utente di aprire una nuova conversazione o riavviare Codex. Se il server MCP continua a non avviarsi, segui la sezione **Se il server MCP non si avvia** e compila con `make build` nella cartella effettiva del plugin installato.
4. Prima di leggere dati, chiama `get_setup_status`. Se `configured` è `false`, segui la procedura della skill: chiama `begin_credentials_setup`, apri il link nel browser laterale di Codex e lascia che l'utente inserisca le credenziali lì. Attendi la sua conferma e verifica di nuovo lo stato.
5. Non chiedere, leggere o copiare le credenziali nella chat o nel repository.

Testo che l'utente può dare al proprio agente:

> Installa il plugin ClasseViva seguendo il README di questo repository. Poi verifica che gli strumenti siano disponibili e, se le credenziali non sono configurate, avvia la procedura nel browser locale. Non chiedermi di scrivere le credenziali in chat.

### Requisiti e piattaforme

Il repository distribuisce un eseguibile già compilato per **Linux amd64**. Se il binario non è compatibile con il sistema o l'architettura dell'utente, oppure il server MCP non si avvia, il plugin può essere compilato localmente dal sorgente con il Makefile incluso. La compilazione richiede Go 1.25 o successivo e `make`.

Se Go non è installato, scegli il sistema operativo e segui la [procedura ufficiale di installazione di Go](https://go.dev/doc/install); poi apri un nuovo terminale e verifica `go version`. Se manca `make`, installa GNU Make usando il gestore di pacchetti o gli strumenti di sviluppo ufficiali del sistema.

## Primo avvio e credenziali

Al primo utilizzo, l'agente controlla lo stato con `get_setup_status`. Se `configured` è `false`, avvia `begin_credentials_setup` e apre una pagina temporanea nel browser laterale di Codex. Inserisci UID e password ClasseViva nella pagina e premi **Salva credenziali**; poi torna alla conversazione e conferma il completamento.

La pagina comunica soltanto con il programma locale su `127.0.0.1`, usa un link casuale e temporaneo e scade dopo 10 minuti. Le credenziali non passano dalla chat. Il percorso esatto del file di configurazione dipende dal sistema e si può ottenere con:

```sh
bin/classeviva --credentials-path
```

Il file contiene l'UID alla prima riga e la password alla seconda. È possibile specificare un percorso diverso con `CLASSEVIVA_CREDENTIALS_FILE`. Su macOS e Linux il programma imposta permessi privati sulla cartella e sul file.

Le credenziali non fanno parte del plugin. Non salvarle nel repository, in `plugin.json`, in `.mcp.json` o nella conversazione. Il file `credentials.txt` è escluso da Git.

## Strumenti MCP

| Strumento | Uso |
| --- | --- |
| `get_setup_status` | Controlla se le credenziali locali sono configurate; non restituisce UID o password. |
| `begin_credentials_setup` | Avvia la pagina temporanea per inserire le credenziali. |
| `get_profile` | Legge il profilo dello studente. |
| `get_agenda` | Legge compiti, lezioni ed eventi per un intervallo di date ISO `YYYY-MM-DD`, inclusivo. Per un giorno usa la stessa data in `from` e `to`. |
| `list_noticeboard` | Elenca gli avvisi senza marcarli come letti. |
| `get_notice_attachment` | Legge un allegato usando `event_code` e `publication_id` dell'avviso. |
| `list_didactics` | Elenca i materiali didattici. |
| `get_didactics_item` | Legge un materiale con il `content_id` restituito da `list_didactics`. |

Tutte le operazioni di ClasseViva sono di sola lettura: il plugin non modifica dati, non segna avvisi come letti e non carica documenti.

## Se il server MCP non si avvia

Se gli strumenti non risultano disponibili anche se il plugin è installato, controlla prima `codex plugin list` e apri una nuova conversazione o riavvia Codex. Se l'errore indica che l'eseguibile è mancante, incompatibile o non può essere avviato, compila il server nella **cartella effettiva del plugin installato**: deve contenere `Makefile`, `go.mod` e `.mcp.json`. Codex usa il binario in `bin/` sotto quella cartella, quindi compilare una copia diversa del repository non aggiorna il plugin installato.

```sh
make build
```

Il comando rileva il sistema e l'architettura locali e produce `bin/classeviva` (oppure `bin/classeviva.exe` su Windows). Se manca Go, segui prima la procedura ufficiale indicata sopra; poi esegui di nuovo `make build`. Dopo la compilazione, riavvia Codex o apri una nuova conversazione per ricaricare il server. Se non parte ancora, riporta l'errore preciso invece di dichiarare disponibili gli strumenti.

## Troubleshooting

- **Gli strumenti non sono disponibili:** controlla l'installazione con `codex plugin list`, poi apri una nuova conversazione o riavvia Codex.
- **Il link di configurazione è scaduto o non valido:** chiedi all'agente di avviare una nuova configurazione e apri l'ultimo link generato da Codex. Il link precedente non è riutilizzabile.
- **Le credenziali risultano mancanti:** controlla soltanto il valore booleano di `get_setup_status` e ripeti la configurazione dal browser. Non stampare né aprire il file delle credenziali.

## Build per chi mantiene il plugin

Per compilare il binario per il sistema e l'architettura correnti:

```sh
make build
```

Il binario Linux amd64 `bin/classeviva` è incluso nel repository per rendere installabile il plugin su Linux senza toolchain Go. Su altri sistemi, `make build` compila un eseguibile nativo a partire dal sorgente; serve Go 1.25 o successivo e `make`.

Per i dettagli sul formato dei plugin Codex, consulta la [documentazione ufficiale OpenAI sui plugin](https://developers.openai.com/plugins/build/plugins).
