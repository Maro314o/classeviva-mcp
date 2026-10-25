---
name: classeviva-setup
description: Usa il plugin ClasseViva per richieste naturali su compiti, agenda, profilo, bacheca, allegati e materiali didattici, anche se l'utente non nomina ClasseViva o la skill; guida inoltre la prima configurazione delle credenziali.
---

# ClasseViva

Usa gli strumenti MCP del plugin per leggere i dati scolastici dell'account configurato localmente. Attiva questa skill anche per richieste come «che compiti ho domani?», «cosa ho per venerdì?» o «ci sono nuovi avvisi?», senza chiedere all'utente di nominare il plugin o la skill.

Se gli strumenti del plugin non sono disponibili nella conversazione corrente, non fingere di aver consultato ClasseViva e non sostituire i dati con supposizioni. Prima verifica che il plugin sia installato e abilitato. Se risulta installato ma il server MCP non si avvia o l'eseguibile è mancante o incompatibile, segui la procedura di compilazione qui sotto prima di concludere che il plugin non è utilizzabile.

## Avvio del server e compilazione locale

Il pacchetto include un binario precompilato per Linux amd64. Se il sistema o l'architettura non sono compatibili, oppure Codex segnala che il processo MCP non può essere avviato:

1. Controlla il sistema operativo e l'architettura dell'utente, quindi verifica che il plugin sia installato. Se gli strumenti non sono caricati nella conversazione corrente, aprine una nuova o riavvia Codex.
2. Individua la cartella effettiva del plugin installato, quella che contiene `.mcp.json`, `Makefile`, `go.mod` e `cmd/classeviva/`. Usa il percorso cache di Codex per il marketplace e la versione installata; non compilare una copia separata del repository, perché il server usa il binario dentro la cartella installata.
3. Verifica che siano disponibili `make` e Go 1.25 o successivo (`go version`). Se Go manca, consulta [le istruzioni ufficiali di installazione di Go](https://go.dev/doc/install), seleziona il sistema e l'architettura corretti, completa l'installazione e apri un nuovo terminale. Se `make` manca, installa GNU Make usando gli strumenti ufficiali del sistema.
4. Dalla cartella del plugin installato, esegui `make build`. Il Makefile compila il sorgente per il sistema e l'architettura correnti, disabilita CGO e sovrascrive il binario atteso dalla configurazione MCP.
5. Riavvia Codex o apri una nuova conversazione e verifica che gli strumenti MCP siano caricati. Se il server continua a non partire, comunica l'errore preciso e non dichiarare di aver letto dati ClasseViva.

Non installare Go da siti di terze parti. Se l'installazione ufficiale richiede privilegi amministrativi non disponibili nell'ambiente, spiega all'utente quali passaggi deve completare e riprendi dopo che Go è disponibile.

## Primo utilizzo e credenziali

Prima di leggere qualsiasi dato, chiama `get_setup_status`.

- Se `configured` è `true`, continua con la richiesta dell'utente.
- Se `configured` è `false`, oppure una chiamata segnala credenziali mancanti o non valide, avvia la configurazione:
  1. Chiama `begin_credentials_setup`.
  2. Apri subito l'URL restituito nel pannello browser di Codex con `mcp__codex_app__open_in_codex`, usando `target: {type: "browser", url: URL}`.
  3. Spiega che l'utente deve inserire UID e password nella pagina locale e premere **Salva credenziali**. Non chiedere mai quei valori nella chat, nei parametri MCP o in un file del progetto.
  4. Attendi che l'utente confermi il completamento. Poi richiama `get_setup_status`; procedi con i dati solo se `configured` è `true`.
- Se il tool per aprire il browser non è disponibile, mostra il link temporaneo e chiedi all'utente di aprirlo nel browser incorporato di Codex.
- Se il link è scaduto, la pagina è stata chiusa o l'utente segnala un errore del link, avvia una nuova sessione con `begin_credentials_setup` e usa il nuovo URL.

Il server della configurazione ascolta solo su `127.0.0.1`; il link è casuale, monouso e scade dopo 10 minuti. Le credenziali vengono salvate localmente dal programma. Non leggere, trascrivere o riportare UID e password; `get_setup_status` restituisce soltanto stato e percorso del file.

## Strumenti

- `get_setup_status`: verifica se esiste una configurazione utilizzabile. Non restituisce credenziali.
- `begin_credentials_setup`: crea il link temporaneo per configurare UID e password nel browser locale.
- `get_profile`: legge il profilo dello studente autenticato.
- `get_agenda`: legge compiti, eventi e lezioni. Richiede `from` e `to` come date ISO `YYYY-MM-DD`, estremi inclusi. Per un solo giorno usa la stessa data in entrambi i campi.
- `list_noticeboard`: elenca gli avvisi senza marcarli come letti.
- `get_notice_attachment`: legge un allegato usando `event_code` e `publication_id` dell'avviso.
- `list_didactics`: elenca i materiali didattici.
- `download_didactics_item`: salva un materiale didattico in un file locale. Richiede `content_id` e il percorso assoluto completo `destination_path`; non sovrascrive file esistenti.

`list_didactics` recupera solo i metadati e non scarica gli allegati. Per sapere quali materiali esistono, i loro nomi o altre informazioni già presenti nell’elenco, usa solo `list_didactics`. Per leggere o analizzare un documento, scaricalo prima con `download_didactics_item` e poi apri il file locale. Non esiste un comando che invii il contenuto del documento all’agente senza salvarlo. Specifica sempre il percorso assoluto completo nel comando; se l’utente non ha indicato dove salvarlo, chiedigli il percorso prima di invocare lo strumento. Se il file esiste già, scegli un altro percorso solo dopo averlo concordato con l’utente.

Per «domani», «oggi» o un giorno della settimana, determina prima la data di calendario nel fuso orario dell'utente, se noto, e passala a `get_agenda` nel formato `YYYY-MM-DD`. Se il fuso orario può cambiare il giorno e non è noto, chiedilo.

## Limiti

Le chiamate a ClasseViva sono di sola lettura: non segnare avvisi come letti, non caricare file e non inviare modifiche al servizio. L'unica scrittura locale è `download_didactics_item`, da usare se l'utente chiede di salvare, leggere o analizzare un documento, sempre nel percorso assoluto indicato. Presenta i risultati con date e materie quando disponibili; distingui chiaramente gli eventuali allegati o file restituiti dagli altri risultati.
