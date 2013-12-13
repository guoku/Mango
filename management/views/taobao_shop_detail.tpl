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
                <td>{{ .Shop.ShopInfo.Nick }}</td>
            </tr>
            <tr>
                <td>店铺信息</td>
                <td>
                    <form class="form-horizontal" method="POST" action="/scheduler/update_taobaoshop_info">
                    <input type="hidden" name="sid" value="{{.Shop.ShopInfo.Sid}}">
                    <div class="control-group">
                        <label class="control-label" for="shop_priority">优先级：</label>
                        <select id="shop_priority" name="priority">
                            {{ $priority := .Shop.CrawlerInfo.Priority }}
                            {{ range .Priorities }}
                            <option {{if eq $priority . }}selected {{end}}> {{ . }}</option>
                            {{ end }}
                        </select>
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_type">类型：</label>
                        <select id="shop_type" name="shoptype">
                            {{ $shop_type := .Shop.ExtendedInfo.Type }}
                            {{ range .TaobaoShopTypes }}
                            <option {{if eq $shop_type . }}selected {{end}}> {{ . }}</option>
                         {{ end }}
                        </select>
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_gifts">礼品：</label>
                            {{ range .Gifts }}
                            <input id="shop_gifts" type="checkbox" name="{{ .Name }}" {{ if  .On }} checked {{end}}/> {{ .Name }} 
                            {{ end }}
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_cycle">周期(小时)：</label>
                        <input id="shop_cycle" name="cycle" type="text" class="input-medium" value="{{ .Shop.CrawlerInfo.Cycle }}">
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_orientational">是否定向：</label>
                        <select id="shop_orientational" name="orientational">
                            <option value="true" {{if .Shop.ExtendedInfo.Orientational }}selected {{end}}>yes</option>
                            <option value="false" {{if not .Shop.ExtendedInfo.Orientational }}selected {{end}}>no</option>
                        </select>
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_singletail">是否原单：</label>
                        <select id="shop_singletail" name="singletail">
                            <option value="true" {{if .Shop.ExtendedInfo.SingleTail }}selected {{end}}>yes</option>
                            <option value="false" {{if not .Shop.ExtendedInfo.SingleTail }}selected {{end}}>no</option>
                        </select>
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="original">是否原创：</label>
                        <select id="shop_original" name = "original">
                            <option value="true" {{if .Shop.ExtendedInfo.Original}}selected{{end}}>yes</option>
                            <option value="false" {{if not .Shop.ExtendedInfo.Original}}selected{{end}}>no</option>
                        </select>
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_commission">是否佣金：</label>
                        <select id="shop_commission" name="commission">
                            <option value="true" {{if .Shop.ExtendedInfo.Commission}} selected{{ end }}>yes</option>
                            <option value="false" {{if not .Shop.ExtendedInfo.Commission}}selected{{end}}>no</option>
                        </select>
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_commission_rate">佣金比例：</label>
                        <input id="shop_commission_rate" name="commission_rate" type="text" class="input-medium" value="{{ .Shop.ExtendedInfo.CommissionRate }}">
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_main_products">主营类别：</label>
                        <input id="shop_main_products" name="main_products" type="text" class="input-medium" value="{{.Shop.ShopInfo.MainProducts}}">
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_credit">卖家信誉：</label>
                        {{.Shop.ShopInfo.ShopScore.Credit}}
                    </div>
                    <div class="control-group">
                        <label class="control-label" for="shop_praise_rate">好评率：</label>
                        {{.Shop.ShopInfo.ShopScore.PraiseRate}}
                    </div>
                    <div class="control-group">
                        <div class ="controls">
                        <button type="submit" class="btn">保存</button>
                        </div>
                    </div>
                                
                    </form>
                </td>
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
                {{ if ne .Title "" }}
                    <td>
                    {{ if .ItemImgs }}
                    <img src="{{index .ItemImgs 0 }}" width="100" height="100"/>
                    {{ end }}
                    </td>
                    <td>{{.Title}}</td>
                    <td><a href="{{.DetailUrl}}">Link</a></td>
                    <td>{{.Price}}</td>
                {{ else }}
                    <td>not ready</td>
                    <td>not ready</td>
                    <td>not ready</td>
                    <td>not ready</td>
                {{ end }}
            </tr>
            {{ end }}
        </table>
    </div>
</div>
<script>
$(document).ready(function(){
        $("form").submit(function(){
               var cf = confirm("确定修改？")
               if(cf){
                $.post("/scheduler/update_taobaoshop_info",$("form").serialize())
               }
            })
        })
</script>
{{ template "paginator.tpl" .}}
