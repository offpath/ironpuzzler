var app = angular.module('adminApp', []);

app.controller('adminCtrl', function ($scope, $http) {
	$scope.refresh = function() {
	    $http.get("admin/api/hunts").success(function (response) {
		    $scope.hunts = response;
		});
	}
    
	$scope.newHunt = function() {
	    $http.get("admin/api/addhunt?name=" + $scope.newName + "&path=" + $scope.newPath)
	    .success(function (response) {
		    $scope.newName = "";
		    $scope.newPath = "";
		    // Umm, yay for eventual consistency, I guess?
		    setTimeout(function () {$scope.refresh();}, 1000);
		});
	}

	$scope.deleteHunt = function(id) {
	    $http.get("admin/api/deletehunt?id=" + id)
	    .success(function (response) {
		    // Umm, yay for eventual consistency, I guess?
		    setTimeout(function () {$scope.refresh();}, 1000);
		});
	}

	$scope.newName = "";
	$scope.newPath = "";

	$scope.refresh();
    });