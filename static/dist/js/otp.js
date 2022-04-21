function loginOTP() {
    console.log("login via otp!")
    hideErrorAlertOTP();

    if ($("#otp-username").val() === "") {
        showErrorAlertOTP("Please enter a otp username");
        return;
    }

    if ($("#otp-passcode").val() === "") {
        showErrorAlertOTP("Please enter a otp passcode");
        return;
    }

    username = $("#otp-username").val()
    passcode = $("#otp-passcode").val()

    $.ajax({
        url: '/verifyTotp',
        type: 'POST',
        data: JSON.stringify({
            username: username,
            passcode: passcode,
        }),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (response) {
            console.log("login success")
            window.location = "/dashboard"
            console.log(response)
        }
    });
}

function hideErrorAlertOTP() {
    $("#otp-alert").hide();
}

function showErrorAlertOTP(msg) {
    $("#otp-alert-msg").text(msg)
    $("#otp-alert").show();
}