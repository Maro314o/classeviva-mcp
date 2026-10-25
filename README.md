# ClasseViva

Plugin locale per Codex con un server MCP in Go. Espone strumenti di sola lettura per la scheda, l'agenda, la bacheca e i materiali didattici ClasseViva.

## Credenziali

Al primo utilizzo, l'agente controlla se le credenziali sono configurate. Se mancano, avvia una pagina temporanea nel browser laterale di Codex: inserisci UID e password lì e premi **Salva credenziali**. Il modulo comunica soltanto con il programma Go sul computer locale e salva automaticamente il file privato `classeviva/credentials.txt` nella directory di configurazione dell'utente. Le credenziali non passano dalla chat.

Il percorso esatto del file si stampa anche con:

```sh
bin/classeviva --credentials-path
```

Il file deve contenere l'UID ClasseViva alla prima riga e la password alla seconda. Su macOS e Linux usare permessi `0600`. È possibile indicare un altro percorso con `CLASSEVIVA_CREDENTIALS_FILE`.

Le credenziali non fanno parte del plugin. Non salvarle nel repository, in `plugin.json`, in `.mcp.json` o nella conversazione con l'agente.

La pagina di configurazione ascolta esclusivamente su `127.0.0.1`, usa un link monouso e scade dopo 10 minuti. Su macOS e Linux il programma imposta permessi `0700` sulla cartella e `0600` sul file.

## Strumenti MCP

- `get_profile`
- `get_agenda` con date ISO inclusive
- `list_noticeboard` senza marcare gli avvisi come letti
- `get_notice_attachment`
- `list_didactics`
- `get_didactics_item`
- `get_setup_status` e `begin_credentials_setup` per la configurazione locale

## Build per chi mantiene il plugin

Il pacchetto Codex include un eseguibile Linux amd64 compilato, quindi chi lo installa non deve installare Go o dipendenze. Per rigenerarlo dal sorgente serve Go:

```sh
go mod download
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o bin/classeviva ./cmd/classeviva
```

La versione iniziale include il binario Linux amd64. Per distribuirla su altri sistemi operativi andranno aggiunti eseguibili compilati per quelle piattaforme.
