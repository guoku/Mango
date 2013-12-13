{{ template "nav.tpl" .}}
<div class="container-fluid">
    <div class="span10">
    <form class="form-search" method="POST" action="/scheduler/add_shop">
        <input type="text" class="input-medium search-query" name="shop_name">
        <button type="submit" class="btn">添加店铺首页URL</button>
    </form>   
    <form class="form-search" method="GET" action="/scheduler/list_shops">
        <input type="text" class="input-medium search-query" name="nick">
        <button type="submit" class="btn">查找店铺昵称</button>
    </form>   
    <div>
        <form class="form-search" method="GET" action="/scheduler/list_shops">
        排序
        <select id="sorting" name="sorton">
            <option value="created_time" {{ if eq .SortOn "created_time" }} selected="selected" {{end}}>创建时间</option>
            <option value="priority" {{ if eq .SortOn "priority" }} selected="selected" {{end}}>优先级</option>
            <option value="status" {{ if eq .SortOn "status" }} selected="selected" {{end}}>状态</option>
            {{$sorton := .SortOn}}
            {{range .Gifts}}
                <option value="{{.}}" {{if eq $sorton .}} selected="selected"{{end}}>{{.}}</option>
            {{ end }}
        </select>
        <button type="submit" class="btn">确定</button>
        </form>
    </div>
    <table class="table table-bordered table-striped">
        <th>Name</th>
        <th>Title</th>
        <th>Priority</th>
        <th>Status</th>
        {{ range .ShopList }}
        <tr>
            <td><a href="/scheduler/shop_detail/taobao/?sid={{ .ShopInfo.Sid }}">{{ .ShopInfo.Nick }}</a></td>
            <td>{{ .ShopInfo.Title }} </td>
            <td>{{ .CrawlerInfo.Priority }}</td>
            <td>{{ .Status }}</td>
        </tr>
        {{ end }}
    </table>
    </div>
</div>
{{ template "paginator.tpl" .}}
