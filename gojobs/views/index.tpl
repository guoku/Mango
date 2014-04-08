{{ template "nav.tpl" . }}

<div class="container-fluid">
    <div class="span10">
        <form class="form-search" method="POST" action="/start">
            <input type="hidden" class="input-medium search-query" name="start" value=1>
            <button type="submit" class="btn">启动</button>
        </form>
    </div>

    <div class="span10">
        <form class="form-search" method="POST" action="/end">
            <input type="hidden" class="input-medium search-query" name="end" value=2>
            <button type="submit" class="btn">停止</button>
        </form>
    </div>
</div>

{{ template "paginator.tpl" .}}
