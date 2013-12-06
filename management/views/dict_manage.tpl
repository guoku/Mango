{{ template "nav.tpl" .}}
<div class="container-fluid">
    <div class="span10">
        <table class="table table-bordered table-striped">
            <th>Word</th>
            <th>Frequency</th>
            <th>Weight</th>
            <th>blacklist</th>
            {{ range .Words }}
                <tr>
                    <td>{{ .Word }}</td>
                    <td>{{ .Freq }}</td>
                    <td>{{ .Weight }}</td>
                    <td>
                        <form method="POST" class="form-inline" action="/scheduler/dict_manage/update/">
                        <input type="hidden" name="w" value="{{.Word}}">
                        {{ if .Blacklisted }}
                        <input type="hidden" name="blacklist" value="false">
                        <button type="submit" class="btn">取消黑名单</button>
                        {{ else }}
                        <input type="hidden" name="blacklist" value="true">
                        <button type="submit" class="btn">加入黑名单</button>
                        {{ end }}
                        </form>
                    </td>
                </tr>
            {{ end }}
        </table>
    </div>
</div>

{{ template "paginator.tpl" .}}
