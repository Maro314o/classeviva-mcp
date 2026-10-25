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
3. Se il plugin non compare nella conversazione corrente, chiedi all'utente di aprire una nuova conversazione o riavviare Codex; poi verifica che siano disponibili gli strumenti ClasseViva.
4. Prima di leggere dati, chiama `get_setup_status`. Se `configured` è `false`, segui la procedura della skill: chiama `begin_credentials_setup`, apri il link nel browser laterale di Codex e lascia che l'utente inserisca le credenziali lì. Attendi la sua conferma e verifica di nuovo lo stato.
5. Non chiedere, leggere o copiare le credenziali nella chat o nel repository.

Testo che l'utente può dare al proprio agente:

> Installa il plugin ClasseViva seguendo il README di questo repository. Poi verifica che gli strumenti siano disponibili e, se le credenziali non sono configurate, avvia la procedura nel browser locale. Non chiedermi di scrivere le credenziali in chat.

### Requisiti e piattaforme

Il repository distribuisce un eseguibile già compilato per **Linux amd64**. Su questa piattaforma non serve installare Go. Le build per macOS, Windows e altre architetture non sono ancora incluse: prima di installare lì il plugin serve compilare e distribuire un eseguibile compatibile con il sistema.

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

## Troubleshooting

- **Gli strumenti non sono disponibili:** controlla l'installazione con `codex plugin list`, poi apri una nuova conversazione o riavvia Codex.
- **Il link di configurazione è scaduto o non valido:** chiedi all'agente di avviare una nuova configurazione e apri l'ultimo link generato da Codex. Il link precedente non è riutilizzabile.
- **Le credenziali risultano mancanti:** controlla soltanto il valore booleano di `get_setup_status` e ripeti la configurazione dal browser. Non stampare né aprire il file delle credenziali.

## Build per chi mantiene il plugin

Per rigenerare il binario Linux amd64 dal sorgente serve Go:

```sh
go mod download
mkdir -p bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o bin/classeviva ./cmd/classeviva
```

Il binario compilato `bin/classeviva` è incluso nel repository per rendere installabile il plugin Linux senza toolchain Go. Per distribuire il plugin su altri sistemi, aggiungi eseguibili compatibili e la relativa configurazione MCP.

Per i dettagli sul formato dei plugin Codex, consulta la [documentazione ufficiale OpenAI sui plugin](https://developers.openai.com/plugins/build/plugins).
