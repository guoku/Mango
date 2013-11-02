{{ template "nav.tpl" .}}
<div class="container-fluid">
    <form class="form-horizontal well" method="POST" action="/user_profile">
        <div class="control-group">
            <p class="text-error">{{.flash.error}}</p>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_email">Email</label>
            <div class="controls">
                <input id="user_email" class="input-medium" value="{{.User.Email}}" readonly="readonly"/>
            </div>
        </div>

        <div class="control-group">
            <label class="control-label" for="user_name">Your Name</label>
            <div class="controls">
                <input id="user_name" name="Name" type="text" class="input-medium" value="{{.User.Name}}"/>
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_nick">Nickname</label>
            <div class="controls">
                <input id="user_nick" name="Nickname" type="text" class="input-medium" value="{{.User.Nickname}}"/>
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_mobile">Mobile</label>
            <div class="controls">
                <input id="user_mobile" name="Mobile" type="text" class="input-medium" value="{{.User.Profile.Mobile}}"/>
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_department">Department</label>
            <div class="controls">
                <select id="user_department" name="Department">
                    <option id="Engineering">Engineering</option>
                    <option id="Marketing">Marketing</option>
                    <option id="Product">Product</option>
                    <option id="Operation">Operation</option>
                    <option id="Other">Other</option>
					{{str2html .Option}}
                </select>
            </div>
        </div>
		<div class="control-group">
			<label class="control-label" for="permission">Permissions</label>
			<div class="controls">
		<table id="permission">
			{{ range .User.Permissions}}
            <tr>
               <td>{{.Name}}</td>
            </tr>
            {{ end }}

		</table>
		</div>
		</div>



        <div class="control-group">
            <div class ="controls">
            <button type="submit" class="btn">修改信息</button>
            </div>
        </div>


    </form>

</div>
 <script type="text/javascript">
	$(document).ready(function(){
    var v = $("#selected").text()

    $("#selected").remove()
	$("#"+v).attr("selected","selected")

});
</script>
