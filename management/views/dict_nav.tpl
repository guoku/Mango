<div class="container">
    <ul class="nav nav-tabs">
        <li {{if eq .TabName "Blacklist" }}class="active"{{end}}><a href="/dict_manage/blacklist">垃圾词管理</a></li>
        <li {{if eq .TabName "Brands"}}class="active"{{end}}><a href="/dict_manage/brands_manage/">品牌词库管理</a></li>
    </ul>
</div>
