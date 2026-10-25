---
name: classeviva-setup
description: Use when configuring this ClasseViva plugin, reading a student's profile, checking homework or agenda, consulting the notice board, or opening didactic materials.
---

# ClasseViva

Questo plugin consulta in sola lettura i dati dell'account ClasseViva configurato localmente.

## Primo utilizzo e configurazione

Prima di leggere dati ClasseViva, chiama `get_setup_status`. Se `configured` è `false`, oppure una chiamata di lettura segnala credenziali mancanti o non valide:

1. Chiama `begin_credentials_setup`.
2. Apri subito l'URL restituito nel pannello laterale del browser di Codex usando `mcp__codex_app__open_in_codex` con `target: {type: "browser", url: URL}`.
3. Spiega all'utente che deve inserire UID e password nella pagina appena aperta e premere **Salva credenziali**. Le credenziali vanno direttamente al programma locale e non devono essere inviate nella conversazione.
4. Attendi che l'utente confermi di aver completato il modulo. Poi richiama `get_setup_status`; se risulta configurato, continua la richiesta originale.

Non chiedere mai all'utente di incollare UID o password nella conversazione, nei parametri di uno strumento o in un file del progetto. Non trascrivere, leggere o riportare i valori inseriti. Se non è disponibile lo strumento per aprire il browser nel pannello, mostra il link temporaneo all'utente e chiedigli di aprirlo nel browser incorporato di Codex; il link scade dopo 10 minuti.

Il server ascolta solo su `127.0.0.1`; il link è casuale e temporaneo. Il modulo salva due righe nel file locale indicato da `get_setup_status`, con permessi privati su macOS e Linux. La pagina si chiude dopo il salvataggio o dopo 10 minuti.

## Strumenti

- `get_setup_status`: controlla la configurazione e restituisce solo lo stato e il percorso locale, mai i valori delle credenziali.
- `begin_credentials_setup`: avvia la pagina web locale temporanea per inserire le credenziali.
- `get_agenda`: richiede `from` e `to` nel formato `YYYY-MM-DD`. Per un giorno specifico usa la stessa data per entrambi.
- `list_noticeboard`: elenca gli avvisi senza marcarli come letti.
- `get_notice_attachment`: usa `event_code` e `publication_id` dell'avviso per leggere un allegato.
- `list_didactics` e `get_didactics_item`: elencano e aprono materiali didattici.
- `get_profile`: restituisce la scheda dello studente autenticato.

Presenta i risultati con date e materie quando disponibili. Le chiamate a ClasseViva sono di sola lettura: non segnare avvisi come letti, non caricare documenti e non inviare modifiche.
