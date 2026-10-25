---
name: classeviva-setup
description: Usa il plugin ClasseViva per richieste naturali su compiti, agenda, profilo, bacheca, allegati e materiali didattici, anche se l'utente non nomina ClasseViva o la skill; guida inoltre la prima configurazione delle credenziali.
---

# ClasseViva

Usa gli strumenti MCP del plugin per leggere i dati scolastici dell'account configurato localmente. Attiva questa skill anche per richieste come «che compiti ho domani?», «cosa ho per venerdì?» o «ci sono nuovi avvisi?», senza chiedere all'utente di nominare il plugin o la skill.

Se gli strumenti del plugin non sono disponibili nella conversazione corrente, spiega che il plugin non è caricato e che va installato o abilitato in Codex. Non fingere di aver consultato ClasseViva e non sostituire i dati con supposizioni.

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
- `get_didactics_item`: legge un materiale usando `content_id` restituito da `list_didactics`.

Per «domani», «oggi» o un giorno della settimana, determina prima la data di calendario nel fuso orario dell'utente, se noto, e passala a `get_agenda` nel formato `YYYY-MM-DD`. Se il fuso orario può cambiare il giorno e non è noto, chiedilo.

## Limiti

Le chiamate a ClasseViva sono di sola lettura: non segnare avvisi come letti, non caricare file e non inviare modifiche. Presenta i risultati con date e materie quando disponibili; distingui chiaramente gli eventuali allegati o file restituiti dagli altri risultati.
