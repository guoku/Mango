{{ template "nav.tpl" .}}

<div class="container-fluid">
    <div class="page-header">
        <h2>User permissions</h2>
    </div>
	<form action="/admin_view" method="post">
    <table class="table table-striped table-bordered table-hover">
        <thead>
            <tr>
                <th>Id</th>
                <th>Name</th>
                <th>Manage password</th>
                <th>Manage crawler</th>
                <th>Manage product</th>

            </tr>
        </thead>
        <tbody>


            <tr>
                <td><input type="hidden" name="Id" value={{.UserData.Id}}>{{.UserData.Id}}</input></td>
                <td><input type="hidden" name="Name" value={{.UserData.Name}}>{{.UserData.Name}}</input></td>
				{{if .UserData.Perm_password}}
					<td><input id="assword" type="checkbox" name="Perm_password" checked="true"/></td>
				{{else}}
					<td><input id="Password" type="checkbox" name="Perm_password"/></td>
				{{end}}
				{{if .UserData.Perm_crawler}}
					<td><input id="Crawler" name="Perm_crawler" type="checkbox"  checked="true"/></td>
				{{else}}
					<td><input id="Crawler" name="Perm_crawler" type="checkbox"  /></td>
				{{end}}
				{{if .UserData.Perm_product}}
					<td><input id="Product" name="Perm_product" type="checkbox"   checked="true"/></td>
				{{else}}
					<td><input id="Product" name="Perm_product" type="checkbox" /></td>
				{{end}}


            </tr>

        </tbody>
    </table>
	<div style="text-align:center">
	<input name="Submit" type="submit" value="提交"/>
	</div>
	</form>
</div>
<script type="text/javascript">
	$(document).submit(function(){

	if($("#Password").attr("checked")=="checked"){
		$("#Password").attr("value","true")

	}
if($("#Crawler").attr("checked")=="checked"){
		$("#Crawler").attr("value","true")
	}
if($("#Product").attr("checked")=="checked"){
		$("#Product").attr("value","true")
	}

});
</script>
