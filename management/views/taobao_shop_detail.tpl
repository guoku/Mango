{{ template "nav.tpl" . }}
<div class="container-fluid">
    <div class="span10">
        <table class="table table-bordered table-striped">
            <th>Taobao ID</th>
            <th>Image</th>
            <th>Title</th>
            <th>Link</th>
            <th>Price</th>
            {{ range .ItemList }}
            <tr>
                <td><a href="/scheduler/item_detail/taobao/{{.NumIid}}">{{.NumIid}}</a></td>
                {{with index .ApiData.ItemImgs.ItemImgArray 0 }}
                    <td><img src="{{.Url}}" width="100" height="100"/></td>
                {{end}}
                <td>{{.ApiData.Title}}</td>
                <td><a href="{{.ApiData.DetailUrl}}">Link</a></td>
                <td>{{.ApiData.Price}}</td>
            </tr>
            {{ end }}
        </table>
    </div>
</div>
{{ template "paginator.tpl" .}}
