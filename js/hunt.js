var app = angular.module('huntApp', ['ngCookies']);

app.controller('huntCtrl', ['$scope', '$cookies', '$http', function ($scope, $cookies, $http) {
	$scope.refreshHunt = function() {
	    $http.get("/api/hunt?hunt_id=" + huntId).success(function (response) {
		    $scope.info = response;
		    $scope.signedIn = ($scope.info.Teams.CurrentTeam != "");
		    if ($scope.info.Teams.BadSignIn) {
			window.alert("Bad password");
			$scope.logout();
		    }
		});
	}

	$scope.login = function() {
	    $cookies.put("team_id", $scope.teamID);
	    $cookies.put("password", $scope.password);
	    $scope.teamID = "";
	    $scope.password = "";
	    $scope.refreshHunt();
	}

	$scope.logout = function() {
	    $cookies.remove("team_id");
	    $cookies.remove("password");
	    $scope.refreshHunt();
	}

	$scope.signedIn = false;
	$scope.teamID = "";
	$scope.password = "";
	$scope.refreshHunt();
    }]);