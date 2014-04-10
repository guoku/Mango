{{ template "nav.tpl" . }}

<div class="container-fluid">
    <table class="table table-bordered table-striped">
        <tr>
            <th>Service Name</th>
            <th>Statu</th>
        </tr>
        <tr>
            {{ range .serviceInfo }}
            <tr>
                <td> <a href="/detail?serviceName={{ .Name }}">{{ .Name }}</a> </td>
                <td>
                <form method="POST" class="narrow-form update_form" margin-bottom="0px" action="/switcher/">
                   <input class="service" type="hidden" name="serviceName" value="{{ .Name }}"/> 
                {{ if compare .Statu "started" }}
                   <button type="submit" class="btn">停止</button>
                <input class="service" type="hidden" name="actionName" value="stop"/>
                {{ else }}
                    <input class="service" type="hidden" name="actionName" value="start"/>
                    <button type="submit" class="btn">启动</button>
                {{ end }}
                </form>
                </td>
            </tr>
            {{ end }}
        </tr>
    </table>
</div>
