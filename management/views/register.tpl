<div class="col-lg-16">
    <form class="form-horizontal well bs-example" method="POST" action="/register">
        <fieldset>
        <legend> 请填写注册信息</legend>
        {{ if .HasErrors }}
            <div class="form-group">
                <div class="col-lg-4 col-lg-offset-2 alert alert-dismissable alert-danger">
                   {{.flash.error}}
                </div>
            </div>
        {{end}}
        <div class="form-group">
            <label class="col-lg-2 control-label" for="user_email">Email</label>
            <div class="col-lg-10">
                <input id="user_email" class="input-medium" value="{{.Invitation.Email}}" readonly="readonly">
            </div>
        </div>
        <div class="form-group">
            <label class="col-lg-2 control-label" for="user_pass">Password</label>
            <div class="col-lg-10">
                <input id="user_pass" name="password" type="password" class="input-medium">
            </div>
        </div>
        <div class="form-group">
            <label class="col-lg-2 control-label" for="user_name">Your Name</label>
            <div class="col-lg-10">
                <input id="user_name" name="name" type="text" class="input-medium">
            </div>
        </div>
        <div class="form-group">
            <label class="col-lg-2 control-label" for="user_nick">Nickname</label>
            <div class="col-lg-10">
                <input id="user_nick" name="nickname" type="text" class="input-medium">
            </div>
        </div>
        <div class="form-group">
            <label class="col-lg-2 control-label" for="user_mobile">Mobile</label>
            <div class="col-lg-10">
                <input id="user_mobile" name="mobile" type="text" class="input-medium">
            </div>
        </div>
        <div class="form-group">
            <label class="col-lg-2 control-label" for="user_department">Department</label>
            <div class="col-lg-2">
                <select id="user_department" name="department" class="form-control">
                    <option>Engineering</option>
                    <option>Marketing</option>
                    <option>Product</option>
                    <option>Operation</option>
                    <option>Other</option>
                </select>
            </div>
        </div>
        <input type="hidden" name="token" value="{{.Invitation.Token}}">
        
        <div class="form-group">
            <div class ="col-lg-10 col-lg-offset-2">
            <button type="submit" class="btn btn-primary">Submit</button>
            </div>
        </div>
        </fieldset>
    </form>
</div>
