var app = angular.module('adminHuntApp', []);

app.controller('ingredientsCtrl', function ($scope, $http) {
	$scope.refresh = function() {
	    $http.get("/admin/api/ingredients?hunt_id=" + huntId).success(function (response) {
		    $scope.ingredients = response;
		    console.log(response);
		});
	}

	$scope.updateIngredients = function() {
	    $http.get("/admin/api/updateingredients?hunt_id=" + huntId +
		      "&ingredients=" + encodeURIComponent($scope.newIngredients));
	    $scope.newIngredients = "";
	    // TODO(dneal): go through the channel
	    setTimeout(function () {$scope.refresh();}, 1000);
	}

	$scope.newIngredents = "";
	$scope.refresh();
    });

app.controller('teamsCtrl', function ($scope, $http) {
	$scope.refresh = function() {
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
			      $scope.refresh();
			  });
	}

	$scope.deleteTeam = function(id) {
	    $http.get("/admin/api/deleteteam?hunt_id=" + huntId +
		      "&team_id=" + id).success(function (response) {
			      $scope.refresh();
			  });
	}

	$scope.newTeamName = "";
	$scope.newTeamPassword = "";
	$scope.newTeamNovice = false;
	$scope.refresh();
    });

app.controller('adminHuntCtrl', function ($scope, $http) {
	$scope.refreshHunt = function() {
	    $http.get("/admin/api/hunt?hunt_id=" + huntId).success(function (response) {
		    $scope.hunt = response;
		    $scope.editable = ($scope.hunt.State == 0);
		    $scope.advanceable = ($scope.hunt.State != 5 && $scope.hunt.State != 7);
		    $scope.hasPuzzles = ($scope.hunt.State != 0);
		    if ($scope.hasPuzzles) {
			$scope.refreshPuzzles();
		    }
		});
	}

	$scope.refreshPuzzles = function() {
	    $http.get("/admin/api/puzzles?hunt_id=" + huntId).success(function (response) {
		    $scope.puzzles = response;
		});
	}
	
	$scope.advanceState = function() {
	    r = window.confirm("Are you sure you wish to advance the state? This cannot be undone.");
	    if (r) {
		$http.get("/admin/api/advancestate?hunt_id=" + huntId +
			  "&currentstate=" + $scope.hunt.State).success(function (response) {
				  $scope.refreshHunt();
			      });
	    }
	}

	$scope.stateTable = {
	    0: "Pre-launch",
	    1: "Novice ingredients released",
	    2: "Ingredients released",
	    3: "Solving",
	    4: "Surveying",
	    5: "Tallying results",
	    6: "Tallying done",
	    7: "Results released",
	}

	$scope.editable = false;
	$scope.advanceable = false;
	$scope.hasPuzzles = false;
	$scope.refreshHunt();
    });