{{ template "nav.tpl" . }}
<div class="container-fluid">
    <div class="span10">
        <table class="table table-bordered table-striped">
            <tr>
                <td>名称</td>
                <td><a href="http://shop{{.Shop.ShopInfo.Sid}}.taobao.com">{{ .Shop.ShopInfo.Title }}</a></td>
            </tr>
            <tr>
                <td>店主昵称</td>
                <td>{{.Shop.ShopInfo.Nick }}</td>
            </tr>
            <tr>
                <td>优先级</td>
                <td>{{ .Shop.CrawlerInfo.Priority }}</td>
            </tr>
            <tr>
                <td>周期(小时)</td>
                <td>{{ .Shop.CrawlerInfo.Cycle }}</td>
            </tr>
        </table>
    </div>
    <div class="span10">
        <table class="table table-bordered table-striped">
            <th>Taobao ID</th>
            <th>Image</th>
            <th>Title</th>
            <th>Link</th>
            <th>Price</th>
            {{ range .ItemList }}
            <tr>
                <td><a href="/scheduler/item_detail/taobao/?id={{.NumIid}}">{{.NumIid}}</a></td>
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
