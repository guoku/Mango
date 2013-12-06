<div class="container">
    <ul class="nav nav-tabs">
        <li {{if .Tab.IsIndex }}class="active" {{end}}>
            <a href="/">首页</a>
        </li>
        <li {{if .Tab.IsPassword }}class="active" {{end}}><a href="/list_pass">密码管理</a></li>
        <li {{if .Tab.IsScheduler }}class="active" {{end}}><a href="/scheduler/list_shops">爬虫管理</a></li>
        <li {{if .Tab.IsCommodity }} class="active" {{end}}><a href="/commodity/category">商品管理</a></li>
        <li {{if .Tab.IsCommodity }} class="active" {{end}}><a href="/scheduler/dict_manage">词库管理</a></li>
		<li {{if .Tab.IsProfile }}class="active" {{end}}><a href="/user_profile">个人信息</a></li>
        <li class="pull-right"><a href="/logout">退出</a></li>
        <li class="pull-right"><h5>{{.User.Name}}</h5></li>
        {{if .User.IsAdmin }}
            <li class="pull-right"><a href="/invite">邀请</a></li>
        {{end}}
    </ul>
</div>
