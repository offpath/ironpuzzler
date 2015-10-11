var app = angular.module('adminHuntApp', []);

app.controller('adminHuntCtrl', function ($scope, $http) {
	$scope.refreshHunt = function() {
	    $http.get("/admin/api/hunt?hunt_id=" + huntId).success(function (response) {
		    $scope.hunt = response;
		    $scope.editable = ($scope.hunt.State == 0);
		    $scope.refreshTeams();
		});
	}

	$scope.refreshTeams = function() {

	}

	$scope.updateIngredients = function(ingredients) {
	    $http.get("/admin/api/updateingredients?hunt_id=" + huntId + "&ingredients=" + encodeURIComponent($scope.hunt.Ingredients));
	}

	$scope.refreshHunt();
    });