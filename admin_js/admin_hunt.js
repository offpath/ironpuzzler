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
	    $http.get("/admin/api/teams?hunt_id=" + huntId).success(function (response) {
		    $scope.teams = response;
		});
	}

	$scope.addTeam = function() {
	    $http.get("/admin/api/addteam?hunt_id=" + huntId +
		      "&name=" + encodeURIComponent($scope.newTeamName) +
		      "&password=" + encodeURIComponent($scope.newTeamPassword) +
		      "&novice=" + $scope.newTeamNovice).success(function (response) {
			      $scope.newTeamName = "";
			      $scope.newTeamPassword = "";
			      $scope.newteamNovice = false;
			      $scope.refreshTeams();
			  });
	}

	$scope.deleteTeam = function(id) {
	    $http.get("/admin/api/deleteteam?hunt_id=" + huntId +
		      "&team_id=" + id).success(function (response) {
			      $scope.refreshTeams();
			  });
	}

	$scope.updateIngredients = function(ingredients) {
	    $http.get("/admin/api/updateingredients?hunt_id=" + huntId +
		      "&ingredients=" + encodeURIComponent($scope.hunt.Ingredients));
	}

	$scope.editable = false;
	$scope.newTeamName = "";
	$scope.newTeamPassword = "";
	$scope.newTeamNovice = false;
	$scope.refreshHunt();
    });