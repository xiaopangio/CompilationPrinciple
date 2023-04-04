import request from "../helper/request"
export const cal = (data) => request({
    method: "post",
    url: "/calculate",
    data
})