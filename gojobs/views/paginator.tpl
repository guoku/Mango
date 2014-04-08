<div class="container">
    <ul class="pager">
        {{ if .Paginator.HasPrev }}
            <li><a href="?p={{ .Paginator.PrevPage }}&{{.Paginator.OtherParams}}">上一页</a></li>
        {{ end }}
        {{ if .Paginator.HasNext }}
            <li><a href="?p={{.Paginator.NextPage}}&{{.Paginator.OtherParams }}">下一页</a></li>
        {{ end }}
    </ul>
</div>
