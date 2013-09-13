{{ template "nav.tpl" .}}
<div class="container-fluid">
    <div class="span10">
    <form class="well form-search" method="POST" action="/scheduler/add_shop">
        <input type="text" class="input-medium search-query" name="shop_name">
        <button type="submit" class="btn">添加店铺昵称</button>
    </form>   
    <table class="table table-bordered table-striped">
        <th>Name</th>
        <th>Title</th>
        <th>Status</th>
        {{ range .ShopList }}
        <tr>
            <td>{{ .ShopInfo.Nick }}</td>
            <td>{{ .ShopInfo.Title }} </td>
            <td>{{ .Status }}</td>
        </tr>
        {{ end }}
    </table>
    </div>
</div>
