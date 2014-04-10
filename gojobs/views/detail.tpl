{{ template "nav.tpl" .}}

<div class="container-fluid">
<div class="span10">
    <table class="table table-bordered table-striped">
        <tr>
            <th>Job Name</th>
            <th>Statu </th>
            <th>Count</th>
        </tr>
        <tr>
            <td>{{ .Info.Name }}</td>
            <td>{{ .Info.Statu }}</td>
            <td>{{ .Info.Count }}</td>
        </tr>
    </table>
    <div class="alert">
        <p id="addresult"></p>
    </div>
    <form class="form-search" method="GET" action="/detail/">
        <input type="text" class="input-medium search-query" name="q"/>
        <input type="hidden" class="input-medium search-query" name="serviceName" value="{{ .Info.Name }}"/>
        <button type="submit" class="btn">类型过滤</button>
    </form>
    <div class="alert">
        <p id="addresult"></p>
    </div>
    <table class="table table-bordered table-striped">
        <tr>
            <th class="span1">日志序号</th>
            <th>日志级别</th>
            <th>日志类型</th>
            <th>错误文件</th>
            <th>错误行数</th>
            <th>错误时间</th>
            <th>错误原因</th>
        </tr>
        {{ range .LogInfo}}
        <tr>
            <td class="span1">{{.Id}}</td>
            <td>{{.Level}}</td>
            <td>{{.LogType}}</td>
            <td>{{.File}}</td>
            <td>{{.Line}}</td>
            <td>{{.Time}}</td>
            <td>{{.Reason}}</td>
        </tr>
        {{ end }}
    </table>
</div>
</div>
{{ template "paginator.tpl" .}}
