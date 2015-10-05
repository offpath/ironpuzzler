var app = angular.module('adminHuntApp', []);

app.controller('adminHuntCtrl', function ($scope, $http) {
	$scope.refresh = function() {
	    $http.get("/admin/api/hunt?hunt_id=" + huntId).success(function (response) {
		    $scope.hunt = response;
		    $scope.editable = ($scope.hunt.State == "Pre-launch")

		});
	}

	$scope.refresh();
    });