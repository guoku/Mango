<div class="container">
    <ul class="nav nav-tabs">
        <li {{if .Tab.IsIndex }}class="active" {{end}}>
            <a href="/">首页</a>
        </li>
        <li {{ if .Tab.IsDetail }}class="active" {{end}}>
            <a href="/detail">详情</a>
        </li>
    </ul>
</div>
