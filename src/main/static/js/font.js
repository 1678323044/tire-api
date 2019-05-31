function getUrlParam(attr){
  var url = location.search;
  var obj = {};
  var arr = url.substr(1).split("&");
  for (var i = 0;i < arr.length;i ++){
    obj[arr[i].split("=")[0] = arr[i].split("=")][1];
  }
  return obj[attr]
}