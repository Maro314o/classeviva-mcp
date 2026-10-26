# ClasseViva

Repository locale per Codex con un server MCP scritto in Go e la skill **ClasseViva**. La skill guida l'agente nell'uso naturale degli strumenti e nella configurazione iniziale. Dopo l'installazione locale puoi chiedere: «Che compiti ho domani?» senza menzionare né il plugin né la skill.

## Installazione locale in Codex

Il plugin usa un server MCP locale. Per renderlo disponibile in tutte le conversazioni locali, clona il repository ed esegui:

```sh
git clone https://github.com/Maro314o/classeviva-mcp
cd classeviva-mcp
make install-local-plugin
```

Il comando compila il server, installa la skill dal marketplace di questo checkout e registra il server MCP con percorso assoluto nel config locale di Codex. In questo modo le nuove conversazioni locali sullo stesso host possono usare `list_lessons` e `list_absences` anche se lavorano in un'altra cartella. Ogni altro utente o host deve eseguire la stessa installazione nel proprio ambiente. Dopo l'installazione, apri una nuova conversazione per caricare gli strumenti.

### Istruzioni per l'agente che esegue l'installazione

Quando l'utente chiede di installare ClasseViva seguendo questo README:

1. Clona il repository `Maro314o/classeviva-mcp` nell'ambiente locale dell'utente.
2. Esegui `make install-local-plugin` dal checkout e controlla `codex plugin list` e `codex mcp list`.
3. Apri una nuova conversazione locale per caricare gli strumenti.
4. Prima di leggere dati, chiama `get_setup_status`. Se `configured` è `false`, segui la procedura della skill: chiama `begin_credentials_setup`, apri il link nel browser laterale di Codex e lascia che l'utente inserisca le credenziali lì. Attendi la sua conferma e verifica di nuovo lo stato.
5. Non chiedere, leggere o copiare le credenziali nella chat o nel repository.

Testo che l'utente può dare al proprio agente:

> Installa il plugin ClasseViva seguendo il README di questo repository. Poi verifica che gli strumenti siano disponibili e, se le credenziali non sono configurate, avvia la procedura nel browser locale. Non chiedermi di scrivere le credenziali in chat.

### Aggiornare un'installazione esistente

Sul computer in cui il plugin è già installato, individua il checkout usato da Codex con `codex plugin marketplace list` e apri quella cartella. Controlla che il checkout sia pulito, poi aggiorna e reinstalla:

```sh
git status --short --branch
git pull --ff-only origin main
make install-local-plugin
```

Infine apri una nuova conversazione locale: quelle già aperte non ricaricano i tool. Il pull aggiorna il clone e `make install-local-plugin` ricompila il server per quel computer e aggiorna plugin e registrazione MCP. Le credenziali restano nella configurazione utente locale e non vengono trasferite o sovrascritte. Non usare reset o pulizie Git per superare modifiche locali. Per la procedura completa, inclusi controlli e casi di conflitto, vedi [la skill locale di aggiornamento](.agents/skills/update-classeviva-mcp/SKILL.md).

### Requisiti e piattaforme

Il repository distribuisce un eseguibile già compilato per **Linux amd64**. Il comando `make install-local-plugin` ricompila il server per aggiornare anche l'installazione locale; la compilazione richiede Go 1.25 o successivo e `make`.

Se Go non è installato, scegli il sistema operativo e segui la [procedura ufficiale di installazione di Go](https://go.dev/doc/install); poi apri un nuovo terminale e verifica `go version`. Se manca `make`, installa GNU Make usando il gestore di pacchetti o gli strumenti di sviluppo ufficiali del sistema.

## Primo avvio e credenziali

Al primo utilizzo, l'agente controlla lo stato con `get_setup_status`. Se `configured` è `false`, avvia `begin_credentials_setup` e apre una pagina temporanea nel browser laterale di Codex. Inserisci UID e password ClasseViva nella pagina e premi **Salva credenziali**; poi torna alla conversazione e conferma il completamento.

La pagina comunica soltanto con il programma locale su `127.0.0.1`, usa un link casuale e temporaneo e scade dopo 10 minuti. Le credenziali non passano dalla chat. Il percorso esatto del file di configurazione dipende dal sistema e si può ottenere con:

```sh
bin/classeviva --credentials-path
```

Il file contiene l'UID alla prima riga e la password alla seconda. È possibile specificare un percorso diverso con `CLASSEVIVA_CREDENTIALS_FILE`. Su macOS e Linux il programma imposta permessi privati sulla cartella e sul file.

Le credenziali non fanno parte del plugin. Non salvarle nel repository, in `plugin.json`, nei file MCP o nella conversazione. Il file `credentials.txt` è escluso da Git.

## Strumenti MCP

Per l'inventario delle rotte ClasseViva effettivamente usate, delle chiamate
diagnostiche e delle API non implementate, consulta
[`docs/classeviva-api.md`](docs/classeviva-api.md).

| Strumento | Uso |
| --- | --- |
| `get_setup_status` | Controlla se le credenziali locali sono configurate; non restituisce UID o password. |
| `begin_credentials_setup` | Avvia la pagina temporanea per inserire le credenziali. |
| `get_profile` | Legge il profilo dello studente. |
| `get_agenda` | Legge compiti, lezioni ed eventi per un intervallo di date ISO `YYYY-MM-DD`, inclusivo. Per un giorno usa la stessa data in `from` e `to`. |
| `list_lessons` | Elenca le lezioni di oggi, di un giorno (`date`) o di un intervallo (`from` e `to`), in formato ISO `YYYY-MM-DD`. |
| `list_absences` | Elenca le assenze complete, da una data (`from`) o in un intervallo (`from` e `to`), in formato ISO `YYYY-MM-DD`. |
| `list_noticeboard` | Elenca gli avvisi senza marcarli come letti. |
| `get_notice_attachment` | Salva l'allegato nel `destination_path` assoluto indicato e restituisce percorso, MIME type e dimensione. Non sovrascrive file esistenti e non segna l'avviso come letto. Usa `event_code`, `publication_id` e `attach_num` (`attachments[].attachNum`; predefinito `1`). |
| `list_didactics` | Elenca i materiali didattici. |
| `download_didactics_item` | Salva il materiale didattico in `destination_path`, un percorso assoluto completo di nome file. Restituisce percorso, MIME type e dimensione; non sovrascrive file esistenti. |

Le chiamate non modificano dati ClasseViva né caricano documenti, ma i due tool di download scrivono sul disco locale nel percorso esplicitamente fornito. Nota: in una verifica API diretta un allegato di un avviso non letto è stato rifiutato con `152:CvvRestApi/access to content not allowed`; dopo la marcatura di lettura lo stesso download è riuscito. Per quel caso la marcatura era necessaria, ma non è una regola universale confermata. Il client MCP nasconde il body degli errori HTTP: l'agente riceve un errore come HTTP 404, non il codice `152`, e un 404 può avere altre cause. Se gli ID verificati nella bacheca sono corretti e il download fallisce, lo stato non letto è una possibile causa, non una certezza. Spiega l'effetto e chiedi consenso prima di qualsiasi marcatura. Il server MCP attuale non espone un tool per segnare un avviso come letto e nessuno dei download lo fa automaticamente. `list_didactics` restituisce solo i metadati; usalo per sapere quali materiali esistono e come si chiamano. Per scaricare un allegato o un materiale didattico, indica un percorso assoluto completo di nome file. I tool non sovrascrivono file esistenti; dopo il download l'agente può aprire il percorso restituito per leggere o analizzare il contenuto.

## Se il server MCP non si avvia

Controlla `codex mcp list` e `codex plugin list`. Se il server non è registrato, oppure punta a un binario vecchio, esegui `make install-local-plugin` dal checkout. Poi apri una nuova conversazione locale. Se il server continua a non partire, riporta l'errore preciso invece di dichiarare disponibili gli strumenti.

## Troubleshooting

- **Gli strumenti non sono disponibili:** controlla `codex plugin list` e `codex mcp list`; dal checkout esegui `make install-local-plugin`, poi apri una nuova conversazione locale.
- **Il link di configurazione è scaduto o non valido:** chiedi all'agente di avviare una nuova configurazione e apri l'ultimo link generato da Codex. Il link precedente non è riutilizzabile.
- **Le credenziali risultano mancanti:** controlla soltanto il valore booleano di `get_setup_status` e ripeti la configurazione dal browser. Non stampare né aprire il file delle credenziali.

## Build per chi mantiene il plugin

Per compilare il binario per il sistema e l'architettura correnti:

```sh
make build
```

Per aggiornare l'installazione locale di Codex dopo modifiche al codice o agli strumenti MCP, esegui:

```sh
make install-local-plugin
```

Il comando registra la sorgente marketplace locale, compila il server, reinstalla la skill e aggiorna il percorso assoluto del server MCP nel config dell'utente. Ripetilo dopo aver aggiornato il checkout per applicare ogni nuova versione. Dopo l'installazione apri una nuova conversazione, perché gli strumenti disponibili nella conversazione corrente non vengono ricaricati automaticamente.

Il binario Linux amd64 `bin/classeviva` è incluso nel repository per rendere installabile il plugin su Linux senza toolchain Go. Su altri sistemi, `make build` compila un eseguibile nativo a partire dal sorgente; serve Go 1.25 o successivo e `make`.

Per i dettagli sul formato dei plugin Codex, consulta la [documentazione ufficiale OpenAI sui plugin](https://developers.openai.com/plugins/build/plugins).
