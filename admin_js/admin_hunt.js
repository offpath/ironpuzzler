var app = angular.module('adminHuntApp', []);

app.controller('adminHuntCtrl', function ($scope, $http) {
	$scope.refreshHunt = function() {
	    $http.get("/admin/api/hunt?hunt_id=" + huntId).success(function (response) {
		    $scope.hunt = response;
		    $scope.editable = ($scope.hunt.State == "Pre-launch");
		    $scope.refreshTeams();
		});
	}

	$scope.refreshTeams = function() {

	}

	$scope.refreshHunt();
    });