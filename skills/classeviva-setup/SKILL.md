---
name: classeviva-setup
description: Usa il plugin ClasseViva per richieste naturali su compiti, lezioni, assenze, agenda, profilo, bacheca, allegati e materiali didattici, anche se l'utente non nomina ClasseViva o la skill; guida inoltre la prima configurazione delle credenziali.
---

# ClasseViva

Usa gli strumenti MCP ClasseViva per leggere i dati scolastici dell'account configurato localmente. Il server MCP è registrato nel config locale di Codex con il nome `classeviva`, così è disponibile in tutte le conversazioni locali e nei subagenti Codex che ereditano i server MCP. Attiva questa skill anche per richieste come «che compiti ho domani?», «quali lezioni ho domani?», «quanti giorni di assenza ho?» o «ci sono nuovi avvisi?», senza chiedere all'utente di nominare il plugin o la skill.

Se gli strumenti MCP ClasseViva non sono disponibili nella conversazione corrente, non fingere di aver consultato ClasseViva e non sostituire i dati con supposizioni. Verifica `codex mcp list` e che il plugin sia installato e abilitato. Dal checkout del repository, `make install-local-plugin` compila il server, registra il percorso assoluto nel config locale di Codex e reinstalla il plugin. Poi apri una nuova conversazione: gli strumenti di una sessione già aperta non si aggiornano.

## Installazione e aggiornamento del server

Il server MCP viene registrato nel config locale di Codex con il nome `classeviva`; il comando contiene il percorso assoluto del binario nel checkout, quindi non dipende dalla cartella del progetto corrente. Per installare o aggiornare server e skill, esegui `make install-local-plugin` dalla radice del repository. Il comando compila il server, aggiorna il marketplace locale, reinstalla il plugin e registra nuovamente il percorso MCP.

Se il server non compare in `codex mcp list`, oppure Codex segnala che il binario è mancante o incompatibile, verifica il sistema operativo e l'architettura, quindi controlla che siano disponibili `make` e Go 1.25 o successivo (`go version`). Se Go manca, consulta [le istruzioni ufficiali di installazione di Go](https://go.dev/doc/install); se manca `make`, installa GNU Make usando gli strumenti ufficiali del sistema. Poi ripeti `make install-local-plugin` e apri una nuova conversazione locale. Non installare Go da siti di terze parti.

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
- `list_lessons`: elenca le lezioni di oggi se non riceve date, di un singolo giorno con `date`, oppure di un intervallo inclusivo con `from` e `to` in formato ISO `YYYY-MM-DD`. Usalo per richieste specifiche sull'orario o sulle lezioni; usa `get_agenda` per la vista combinata di compiti ed eventi.
- `list_absences`: elenca i dettagli delle assenze. Senza date restituisce tutti i dati; `from` da solo restituisce i dati a partire da quella data; `from` e `to` insieme delimitano un intervallo inclusivo. Per un giorno usa la stessa data in entrambi i campi.
- `list_noticeboard`: elenca gli avvisi senza marcarli come letti.
- `get_notice_attachment`: scarica l'allegato nel percorso assoluto completo `destination_path`. Richiede `event_code`, `publication_id` e `destination_path`; `attach_num` proviene da `attachments[].attachNum` e predefinito a `1`. Restituisce percorso, MIME type e dimensione; non sovrascrive file e non segna l'avviso come letto.
- `list_didactics`: elenca i materiali didattici.
- `download_didactics_item`: scarica e salva il materiale didattico nel percorso assoluto completo `destination_path`. Richiede `content_id` e `destination_path`; restituisce percorso, MIME type e dimensione, e non sovrascrive file esistenti.

`list_didactics` recupera solo i metadati e non scarica gli allegati. Per sapere quali materiali esistono, i loro nomi o altre informazioni già presenti nell’elenco, usa solo `list_didactics`. Per scaricare, leggere o analizzare un documento, usa `download_didactics_item` indicando un percorso assoluto completo, poi apri il file locale restituito. Se l’utente non indica dove salvarlo, chiedi il percorso prima della chiamata. Il tool non sovrascrive file esistenti.

Per «domani», «oggi» o un giorno della settimana, determina prima la data di calendario nel fuso orario dell'utente, se noto, e passala a `list_lessons` quando la richiesta riguarda le lezioni, a `list_absences` quando riguarda le assenze, oppure a `get_agenda` quando riguarda compiti o eventi. Usa il formato `YYYY-MM-DD`. Se il fuso orario può cambiare il giorno e non è noto, chiedilo.

## Limiti

Gli strumenti MCP non modificano dati ClasseViva, non segnano avvisi come letti e non caricano file; i due tool di download salvano invece il file sul disco locale nel percorso assoluto indicato. Nota sugli allegati: in una verifica API diretta, il download di un avviso non letto ha restituito `152:CvvRestApi/access to content not allowed`; lo stesso download è riuscito dopo la marcatura di lettura. Per quel caso la marcatura era necessaria, ma non è confermata come regola universale. Il client MCP nasconde il corpo delle risposte HTTP non riuscite, quindi l'agente vedrà un errore di stato, per esempio HTTP 404, e non il codice `152`; un 404 può avere anche altre cause. Se gli identificativi copiati dai metadati sono corretti e il download fallisce, spiega che lo stato non letto è una possibile causa, senza presentarla come certa. Chiedi consenso esplicito prima di qualsiasi marcatura. Il server MCP attuale non espone un tool per segnare un avviso come letto: non farlo in autonomia e non dichiarare che il tool di download lo abbia fatto. Presenta percorso, MIME type e dimensione e apri il file locale restituito se l'utente chiede di leggere o analizzare il contenuto. Presenta i risultati con date e materie quando disponibili.
