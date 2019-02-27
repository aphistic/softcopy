<html>
    <head>
        <title>Importers</title>
    </head>
    <body>
        Importers<br />
        <ul>
            {{ range . }}
                <li>
                    <a href="/importers/{{ .Name  }}">{{ .Name }}</a>
                </li>
            {{ end }}
        </ul>
    </body>
</html>